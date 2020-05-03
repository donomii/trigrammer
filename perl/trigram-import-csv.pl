use strict;
use Data::Dumper;

use DBI;
use lib '.';
use Tri;

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

sub importCSV {
    my $db = shift;
    my $filename = shift;
    open(my $fh, '<:encoding(UTF-8)', $filename)
    or die "Could not open file '$filename' $!";

    my $headers = <$fh>;
    chomp $headers;
    my @headers = split /\,/, $headers;
    while (my $row = <$fh>) {
      chomp $row;
      #print "$row\n";
      my @vals = split /\,/, $row;
      my %h;
      @h{@headers} = @vals;
      Tri::insertHash($db, \%h);
    }
    close $fh;
}


my $fname = shift @ARGV;
my $csvname = shift @ARGV;
unless ($fname && $csvname) {
    print "trigram-import-csv

Use:    trigram-import-csv.pl database.sqlite data.csv

Imports a csv into a trigram database, making trigrams from the columns and values of the csv file.··

The csv must have the column names as the first row.
";
exit;
}
my $db = openDB($fname);
importCSV($db, $csvname);
