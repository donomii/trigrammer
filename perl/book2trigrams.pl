use strict;
use Data::Dumper;
 use Text::CSV;


 my $csv = Text::CSV->new ( { binary => 1 } )  # should set binary attribute.
                 or die "Cannot use CSV: ".Text::CSV->error_diag ();
sub storeTri {
    $csv->print(*STDOUT, \@_);
    print "\n";
}

my $book = join(" ", <STDIN>);
my @words = split /\s+|\.|,|\"/, $book;
foreach (@words) { s/\s+//g; };
@words = grep { $_ } @words;
for my $i (0..scalar(@words)) {
    storeTri($words[$i], $words[$i+1], $words[$i+2]);
}
