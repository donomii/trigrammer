package Tri;
use strict;
use Data::Dumper;

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



my $count=0;


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

1;
