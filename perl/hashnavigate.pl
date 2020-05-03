use strict;

use DBI;

$|++;
my %sym;
my @syml;
my $sym_ind =1;
my %sym_cache;
my %string_cache;

use DBI;
use strict;

sub openDB {
    my $filename = shift;
    my %h;
    my $driver   = "SQLite"; 
    my $database = $filename;
    my $dsn = "DBI:$driver:dbname=$database";
    my $userid = "";
    my $password = "";
    my $dbh = DBI->connect($dsn, $userid, $password, { RaiseError => 1 }) or die $DBI::errstr;

    $dbh->do("PRAGMA synchronous = OFF;PRAGMA journal_mode = WAL;");
    $dbh->do("CREATE TABLE IF NOT EXISTS trigrams ( a string, b string, c string, UNIQUE (a,b,c) ON CONFLICT IGNORE);");
    #$dbh->do("CREATE TABLE IF NOT EXISTS trigrams ( a string, b string, c string );");
    $dbh->do("CREATE TABLE IF NOT EXISTS trigram_symbols( a INT, b INT, c INT, UNIQUE (a,b,c) ON CONFLICT IGNORE);");
    $dbh->do("CREATE TABLE IF NOT EXISTS quadgram_symbols( a INT, b INT, c INT, d INT, UNIQUE (a,b,c,d) ON CONFLICT IGNORE);");
    $dbh->do("CREATE TABLE IF NOT EXISTS strings ( id INTEGER PRIMARY KEY AUTOINCREMENT, s string, UNIQUE (s) ON CONFLICT IGNORE);");

    $dbh->{AutoCommit} = undef;


    my $insert_stmt = qq(INSERT OR IGNORE INTO TRIGRAMS (a,b,c) VALUES (?,?,?););
    $h{insert_sth} = $dbh->prepare( $insert_stmt );

    my $insert_tsyms_stmt = qq(INSERT OR IGNORE INTO TRIGRAM_SYMBOLS (a,b,c) VALUES (?,?,?););
    $h{insert_tsyms_sth} = $dbh->prepare( $insert_tsyms_stmt );

    my $insert_qsyms_stmt = qq(INSERT OR IGNORE INTO QUADGRAM_SYMBOLS (a,b,c,d) VALUES (?,?,?,?););
    $h{insert_qsyms_sth} = $dbh->prepare( $insert_qsyms_stmt );

    my $insert_string_stmt = qq(INSERT OR IGNORE INTO strings (s) VALUES (?););
    $h{insert_string_sth} = $dbh->prepare( $insert_string_stmt );

    my $fetch_string_stmt = qq(SELECT s from strings WHERE id=?;);
    $h{fetch_string_sth} = $dbh->prepare( $fetch_string_stmt );

    my $fetch_sym_stmt = qq(SELECT id from strings WHERE s=?;);
    $h{fetch_sym_sth} = $dbh->prepare( $fetch_sym_stmt );

    $h{database} = $database;
    $h{handle} = $dbh;
    return \%h;
}

sub insert_string {
    my $db = shift;
    my $s = shift;
    $db->{insert_string_sth}->execute($s);
}

sub fetch_string {
    my $db = shift;
    my $id = shift;
    my $str = $sym_cache{$id};
    if ($str) {
        return $str;
    } else {
        #print "Fetching string for $id\n";
        $db->{fetch_string_sth}->execute($id);
        my $res = $db->{fetch_string_sth}->fetchall_arrayref();
        $str =  $res->[0]->[0];
        $string_cache{$id} = $str;
    }
    return $str;
}

sub fetch_sym {
    my $db = shift;
    my $id = shift;
    my $sym = $sym_cache{$id};
    if ($sym) {
        return $sym;
    } else {
        #print "Fetching sym for $id\n";
        $db->{fetch_sym_sth}->execute($id);
        my $res = $db->{fetch_sym_sth}->fetchall_arrayref();
        $sym =  $res->[0]->[0];
        $sym_cache{$id} = $sym;
    }
    return $sym;
}

sub get_create_symbol {
    my $db = shift;
    my $str = shift;
    insert_string($db, $str);
    return fetch_sym($db, $str);
}

sub storeTri {
    my $db = shift;
    my @tri = @_;
    $db->{insert_tsyms_sth}->execute(map {get_create_symbol($db, $_)} @tri);
}


sub storeQuad {
    my $db = shift;
    my @quad = @_;
    $db->{insert_qsyms_sth}->execute(map {get_create_symbol($db, $_)} @quad);
}



print "Opened database successfully\n";

if (0) {
    my $db = openDB("book.sqlite");
    my $book = join(" ", `cat pg1661.txt`);
    my @words = split /\s+|\.|,|\"/, $book;
    foreach (@words) { s/\s+//g; };
    @words = grep { $_ } @words;
    #print @words;
    for my $i (0..scalar(@words)) {
        storeTri($db, $words[$i], $words[$i+1], $words[$i+2]);
    }
    $db->{handle}->commit;


    my $s = "\"You could not possibly have gone at a better time, my dear Watson,\" he said cordially.";
    my @words = split /\s+|\.|,|\"/, $s;
    foreach (@words) { s/\s+//g; };
    @words = grep { $_ } @words;
    print @words;

    my @score;
    for my $i (0..scalar(@words)) {
        my ($aa, $bb, $cc) = map { get_create_symbol($db, $_)} ($words[$i], $words[$i+1], $words[$i+2]);
        #print "Searching for $aa,$bb,$cc\n";
        my $stmt = qq(SELECT DISTINCT c FROM trigram_symbols WHERE a=? AND b=? AND c=?);
        my $sth = $db->{handle}->prepare( $stmt );
        $sth->execute($aa, $bb, $cc);
        my $results = $sth->fetchall_arrayref();
        #print Dumper($results);
        my @results = map { $_->[0] } @$results;
        if (@results) {
            $score[$i]++;;
            $score[$i+1]++;;
            $score[$i+2]++;;
        }
    }
    for my $i (0..scalar(@words)-2) {
        print "'".$words[$i]."'";
        print $score[$i];
        print " ";
    }

    print "\n";
}

use Data::Dumper;
my $count=0;


if (!@ARGV) {
    warn "\nUse:  hexplore data.csv more.csv\n\n";
    warn "Interactive mode\n";
} else {
my $db = openDB("trigrams.sqlite");
    warn "Importing files: @ARGV\n";
    foreach my $file (@ARGV) {
        print "Loading $file...";
        importCSV($db, $file);
        print "done!\n";
    }
}

sub importCSV {
    my $db = shift;
    my $filename = shift;
    open(my $fh, '<:encoding(UTF-8)', $filename)
    or die "Could not open file '$filename' $!";

    my $headers = <$fh>;
    chomp $headers;
    my @headers = split /\|/, $headers;
    while (my $row = <$fh>) {
      chomp $row;
      #print "$row\n";
      my @vals = split /\|/, $row;
      my %h;
      @h{@headers} = @vals;
      insertHash($db, \%h);
    }
    close $fh;
}

my @top;


sub insertHash {
    my $db = shift;
    my $r = shift;
    foreach my $e (keys %$r) {
        my $expanded_rec = {};
        my $string_r = Dumper($r);
        foreach my $ee (keys %$r) {
            next if $e eq $ee;
                my $f = lc $e;
                my $ff = lc $ee;
                my $valf = lc $r->{$f};
                my $valff = lc $r->{$ff};
            #print "$ee: $c\n";
                #$expanded_rec->{$r->{$ee}} = $ee;
                #$expanded_rec->{$ee} = $r->{$ee};


                my $record = $r;
                $record = 1;

                #value -> key -> value
                #push @{$t{$valf}->{$ff}->{$valff}}, 1;
                storeTri($db, $valf, $ff, $valff);
                storeQuad($db, $valf, $ff, $valff, $string_r);
                #push @{$t{$valff}->{$ff}->{$valf}}, 1;

                #key -> value -> key
                #push @{$t{$ff}->{$valff}->{$f}}, 1;
                storeTri($db, $ff, $valff, $f);
                storeQuad($db, $ff, $valff, $f, $string_r);
                #push @{$t{$f}->{$valff}->{$ff}}, 1;

        }
        #push @{$s{$r->{$e}}}, $expanded_rec;
        #push @{$s{$e}}, $expanded_rec;
        $db->{handle}->commit;
    }
    #print Dumper(\%s);
    if ($count++ % 100 == 0) { print "."; }
}

warn "Top: ".scalar(@top)." entries\n";



#foreach my $r (@{$dbh->selectall_arrayref("select * from table_2",{ Slice => {} })}) {
#    foreach my $e (keys %$r) {
#        foreach my $ee (keys %$r) {
#            #print "$ee: $c\n";
#            $s{$r->{$e}}->{$r->{$ee}} = $ee;
#        }
#    }
#    #print Dumper(\%s);
#}





sub recurse_cats {
    my ($input, $path, @cats) = @_;
    return unless $input;
    return unless @cats;
    return unless (ref($input) eq 'ARRAY');
    foreach my $res (@$input) {
    while (my ($key,$value) = each %$res) {
        if ($value eq $cats[0]) {
            if (scalar(@cats) == 1) {
                print join("/", @$path).  "/". $key."\n";
            }
            #my $entries = $s{$key};
            #recurse_cats($entries, [@$path, $key], @cats[1..(scalar(@cats) -1)]);
            }
        }
    }

}

my $last = "";

sub setSearch {
while (my $a = <>) {
    chomp $a;
    my ($cmd, $rest) = split / /, $a, 1;
    if ($cmd eq "dump") {
        #my $res = $s{$last};
        #print Dumper($res);
    }
    my ($data, @cats) = split /\//, $a;
    $last = $data;
    #my $res = $s{$data};
    if (@cats) {
        #recurse_cats($res, [$data],  @cats);
    } else {
        #print Dumper($res);
    }
    print "'dump' to print last result in full\n";
    print "Query> ";
}


    while (my $a = <>) {
        chomp $a;
        my ($cmd, $rest) = split / +/, $a, 1;
        if ($cmd eq "dump") {
            #my $res = $s{$last};
            #print Dumper($res);
        } elsif ($cmd eq 'q') {
            my (@chunks) = split /->/, $a;
            #recurse_chunks(\@chunks, \%s);
        } else {
            print "Unrecognised command\n";
        }
        print "'dump' to print last result in full\n";
        print "'q x/y->z/a' to query\n";
        print "Query> ";
    }
}

sub recurse_chunks {
    my $chunks = shift;
    my $data = shift;
    my $current = $chunks->[0];
    my @rest = @$chunks[1,scalar(@$chunks)-1];
    my %results;
    if (@$chunks) {
            my ($a, $b,$c) = split /\//, $current;
            
            }
}




use List::Util qw[ min ];
sub sampleKeys {
    my $hash = shift;
    my @keys = keys %$hash;
    @keys = grep { $_ ne '' } @keys;
    my $take = min(8, scalar(@keys));
    return @keys[0..$take-1] ;
}

sub sampleDbResults {
    my $keys = shift;
    my @keys = grep { $_ ne '' } @$keys;
    my $take = min(8, scalar(@keys));
    return @keys[0..$take-1] ;
}


sub formatKeys {
    my @keys = @{$_[0]};
    @keys = map { "[".(length($_)>15?substr($_, 0,15)."...":$_)."]" } @keys;
    return @keys;
}

sub arrayToHash {
    my ($key, $arr) = @_;
    my $ret = {};
    foreach my $e (@$arr) {
        $ret->{$e->{$key}} = $e;
    }
    return $ret;
}


my $a;
my $b;


sub shellHash {
use Term::ReadLine;
  my $term = Term::ReadLine->new('Hash Explorator');
  my $prompt = "Ready> ";
  my $OUT = $term->OUT || \*STDOUT;
  #my @keys = sampleKeys($currhash);
  #print $OUT "Here: @keys \n";
  while ( defined ($_ = $term->readline($prompt)) ) {
    chomp;
    $a = $b;
    $b = $_;
    #print Dumper($s{$a});
    #$currList = $s{$a};
    #my $resultHash = arrayToHash($b, $currList);
    #my @keys = sampleKeys($resultHash);
    #print $OUT "@$currList\n";
    #if (scalar(@{$currList}) == 1) {
        #@keys = sampleKeys($currList->[0]);
        #print $OUT "Leaf($a, $b): @{[$resultHash]}\n";
    #} else {
        #print $OUT $a.'->'.$b.": @keys \n";
    #}
    #my $ret = $currhash->{$_};
    #if ($ret) {


        #my @candidates = @{$ret};
        #if (@candidates) {
            #$currhash = arrayToHash($_, \@candidates);
        #}

        #my $candidate = $candidates[0];
        #print $OUT "$_ -> $candidate\n";
        #my $ret = $s{$newKey};
        #print Dumper($candidate);
        #$currhash = $candidate if defined($candidate) && $candidate && ref($candidate) eq "HASH";
        #$term->addhistory($_) if /\S/;
    #}
  }
}

my $currList = [];

use Data::Printer;


my $prev = undef;

sub getBySingle {
    my ($db, $aa) = @_;
    my $stmt = qq(SELECT DISTINCT b FROM trigram_symbols WHERE a=?);
    my $sth = $db->{handle}->prepare( $stmt );
    $sth->execute($aa);
    my $results = $sth->fetchall_arrayref();
    my @results = map { fetch_string($db, $_->[0]) } @$results;
    return \@results;
}


sub getByPair {
    my ($db, $aa, $bb) = @_;
    my $stmt = qq(SELECT DISTINCT c FROM trigram_symbols WHERE a=? AND b=?);
    my $sth = $db->{handle}->prepare( $stmt );
    $sth->execute($aa, $bb);
    my $results = $sth->fetchall_arrayref();
    my @results = map { fetch_string($db, $_->[0]) } @$results;
    return \@results;
}

sub getByTriple {
    my ($db, $aa, $bb, $cc) = @_;
    my $stmt = qq(SELECT DISTINCT d FROM quadgram_symbols WHERE a=? AND b=? AND c=?);
    my $sth = $db->{handle}->prepare( $stmt );
    $sth->execute($aa, $bb, $cc);
    my $results = $sth->fetchall_arrayref();
    my @results = map { fetch_string($db, $_->[0]) } @$results;
    return \@results;
}


my $currList=[];
sub trigrams {
    my $db = shift;
    use Term::ReadLine;
    use Term::ReadLine::Zoid;
  my $term = Term::ReadLine->new('Trigram Explorator');
  my $prompt = "Ready> ";
  my $OUT = $term->OUT || \*STDOUT;
  my @keys = formatKeys([@$currList]);
  print $OUT "Start: @keys"."\n";
  while ( defined ($_ = $term->readline($prompt)) ) {
    chomp;
    s/^\s+//;
    s/\s+$//;
    if ( /^exit$/ || /^quit$/ ) { exit(0);}
    unless ($_) {
        my $dbResult = getByPair($db, $a, $b);
        if (@$dbResult) {
            print Dumper($dbResult);
            if (defined($dbResult) && scalar(@$dbResult)==1) {
                my @keys = @$dbResult;
                $_ = $keys[0];
            } else {
                next;
            }
        }
    }
    print "\n";
    if ($_ eq 'dump') {
        print "Dumping\n";
        if (defined($prev) && defined($a) && defined($b)) {
            my $dbResult = getByTriple($db, $prev, $a, $b);
            if (@$dbResult) {
                p $dbResult;
            }
        }
    } else {
        $prev = $a;
        $a = $b;
        $b = $_;
        $prompt = "$prev -> $a -> $b > ";
        if (defined($a) && defined($b)) {
            my $dbResult = getByPair($db, $a, $b);
            if (@$dbResult) {
                #my @keys = formatKeys([sampleKeys($resultHash)]);
                print Dumper($dbResult);
                my @keys = formatKeys([sampleDbResults($dbResult)]);
                print $OUT $prev.'->'.$a.'->'.$b.": @keys \n";
                set_completion_function(@$dbResult);
            } else {
                doSingleQuery($db, $b);
            }
        } else {
            doSingleQuery($db, $b);
        }
      }
    }
}

sub doSingleQuery {
    my ($db, $b) = @_;
            my $dbResult = getBySingle($db, $b);
            my @keys = formatKeys([sampleDbResults($dbResult)]);
            print "$b: @keys\n";
            set_completion_function(@$dbResult);
}

sub set_completion_function {
    my @keys = @_;
    $readline::rl_completion_function = sub {
        my ($word, $buffer, $start) = @_;
        return @keys unless $word;
        my @cand = grep { $_ =~ m/^$word/ } @keys;
        return @cand;
        };


}




#shellHash();
my $db = openDB("trigrams.sqlite");
trigrams($db);
