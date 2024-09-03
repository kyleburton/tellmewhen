use strict;
use warnings;
use File::stat;
use Time::localtime;

my $fh = $ARGV[0];
print "STAT on fh=$fh\n";
my $timestamp = stat($fh)->mtime;
printf "$timestamp\n";
