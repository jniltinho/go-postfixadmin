#!/usr/bin/perl
use strict;

use lib "/usr/local/nts/libexec/";
use HTTP::Date;

my $delEmail = shift;

if (! $delEmail) {
	print STDERR "USO:\n";
	print STDERR "queue_clean.pl <email_to_delete>\n";
	print STDERR "email default: MAILER-DAEMON\n";
	$delEmail = "MAILER-DAEMON";

}

$_ = $delEmail;
s/\@/\\@/g;
$delEmail = $_;

my $POSTQUEUE = $sys_conf{'postqueue'};
my $POSTSUPER = $sys_conf{'postsuper'};

open(MAILQ, "$POSTQUEUE -p|") || die "postqueue error \n";
my $delta = 0;
my $x = 0;
my $sender = "";
my $recipient = "";
my %Q;
my $queueid;
my $size;
my $date;
my $time;
my $error;
my $recipient;
my $totalsize;
my $totalrequests;

while (<MAILQ>) {
    chomp;
    if (/^Mail queue is empty/) { print "Nada na fila\n"; exit ; }
    

# if ( /^([0-9A-F]{8,16})[\* !] *(\d+).+:\d{2} +(.+)/ ) { 
    if ( /^([0-9A-F]*)[ \*\!]*(\d+) *([a-zA-Z]{3} [a-zA-Z]{3} [ 0-9]{2} \d{2}:\d{2}:\d{2}) +([^ ]+)/ ) {
                $queueid = $1; $size = $2; $date = $3; $sender = $4; 
                $date = str2time($date);
		$time = time;
		$delta = $time - $date;
		$error = ""; $recipient = ""; 
        } elsif ( /^ *\((.+)\)/ ) { 
                $error = $1; 
        } elsif ( /^ *(.+.+)/ ) { 
                $recipient = $1; 
        } elsif ( /^-- (\d+) Kbytes in (\d+) Requests./ ) { 
                $totalsize = $1; 
                $totalrequests = $2; 
        } elsif ( /^-Que.*/ ) { 
                # do nothing 
        } elsif (($delta > 1) and ("$recipient" ne "")) {
		# print $queueid . " " . $delta . " " . $sender . " " . $recipient . "\n";
		if ( ($sender =~ m/$delEmail/) or 
		     ($recipient =~ m/$delEmail/) )
		 {
		    $Q{$queueid} = 1;
		    $x++;
		    # $result = `$POSTSUPER -d $queueid`;
		    # print "delete $recipient $queueid:" . "$result" . "\n";
		    }
		$delta = 0;
		$sender = "";
		$recipient = "";
	} elsif (($delta > 0) and ("$recipient" ne "")) {
	    $delta = 0;
	    $sender = "";
	    $recipient = "";
	}


}
close(MAILQ);

if ( $x eq 0 ) { print "nothing to do\n" ; exit 0; } ;
open(POSTSUPER, "|$POSTSUPER -d -") || die "postsuper error \n";

foreach (keys %Q) {
    print POSTSUPER "$_\n";
}

close(POSTSUPER);
    
