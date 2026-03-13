#!/usr/bin/perl
# filter.pl — Postfix content filter (Perl version)
use strict;
use warnings;
use POSIX qw(strftime);
use Sys::Syslog qw(:standard :macros);

use constant {
    EX_OK       => 0,
    EX_TEMPFAIL => 75,
    INSPECT_DIR => '/var/spool/filter',
    SENDMAIL    => '/usr/sbin/sendmail',
};

my ($stress, $size, $host_ip, $host_name, $helo, $from) = ('', 0, '', '', '', '');
my @recipients;

# --------------------------------------------------------------------------
sub usage {
    my $prog = $0;
    print <<"END";
Usage: $prog [OPTIONS] -- RECIPIENT [RECIPIENT ...]

Postfix content filter: saves the message to a temp file,
logs it to syslog (mail.info) and re-injects via sendmail.

Options:
  --from      <addr>   Envelope sender address (lowercased)
  --host-ip   <ip>     Client IP address
  --host-name <name>   Client hostname
  --helo      <name>   Client HELO/EHLO name
  --size      <bytes>  Message size hint from Postfix
  --stress    <value>  Postfix stress flag
  --                   Start of recipient list
  --help               Show this help and exit

Paths:
  Temp dir : @{[INSPECT_DIR]}
  Sendmail : @{[SENDMAIL]}

Exit codes:
  0   Success
  75  Temporary failure (Postfix will retry)

Postfix master.cf configuration:

  # Step 1: redirect smtp input through the filter
  smtp      inet  n  -  y  -  -  smtpd
    -o content_filter=filter:dummy

  # Step 2: declare the filter service
  filter    unix  -  n  n  -  10  pipe
    flags=Rq user=filter argv=/usr/local/bin/$prog --stress 0 --size \${size} --host-ip \${client_address} --host-name \${client_name} --helo \${client_helo} --from \${sender} -- \${recipient}
END
}

# --------------------------------------------------------------------------
sub parse_args {
    my @args = @_;
    my $rcpt = 0;
    for (my $i = 0; $i < @args; $i++) {
        if ($rcpt) {
            my $r = $args[$i];
            $r =~ s/^\s+//;
            push @recipients, $r if $r ne '';
            next;
        }
        if    ($args[$i] eq '--help')      { usage(); exit EX_OK; }
        elsif ($args[$i] eq '--')          { $rcpt = 1; }
        elsif ($args[$i] eq '--stress'    && $i+1 < @args) { $stress    = $args[++$i]; }
        elsif ($args[$i] eq '--size'      && $i+1 < @args) { $size      = int($args[++$i]); }
        elsif ($args[$i] eq '--host-ip'   && $i+1 < @args) { $host_ip   = $args[++$i]; }
        elsif ($args[$i] eq '--host-name' && $i+1 < @args) { $host_name = $args[++$i]; }
        elsif ($args[$i] eq '--helo'      && $i+1 < @args) { $helo      = $args[++$i]; }
        elsif ($args[$i] eq '--from'      && $i+1 < @args) { $from      = lc($args[++$i]); }
    }
}

# --------------------------------------------------------------------------
parse_args(@ARGV);

openlog('postfix/filter', 'pid', LOG_MAIL);

unless (@recipients) {
    syslog(LOG_ERR, 'FILTER_ERROR: no recipients');
    closelog();
    exit EX_TEMPFAIL;
}

unless (-d INSPECT_DIR) {
    mkdir(INSPECT_DIR, 0750) or do {
        syslog(LOG_ERR, 'FILTER_ERROR: cannot create %s: %s', INSPECT_DIR, $!);
        closelog();
        exit EX_TEMPFAIL;
    };
}

# Temp file: /var/spool/filter/<timestamp>_<pid>.eml
my $tmp = sprintf('%s/%d_%d.eml', INSPECT_DIR, time(), $$);

open(my $fh, '>', $tmp) or do {
    syslog(LOG_ERR, 'FILTER_ERROR: cannot create temp file %s: %s', $tmp, $!);
    closelog();
    exit EX_TEMPFAIL;
};

my $nbytes = 0;
while (my $chunk = do { local $/ = \65536; <STDIN> }) {
    print $fh $chunk or do {
        syslog(LOG_ERR, 'FILTER_ERROR: cannot save mail: %s', $!);
        close($fh);
        unlink($tmp);
        closelog();
        exit EX_TEMPFAIL;
    };
    $nbytes += length($chunk);
}
close($fh);

syslog(LOG_INFO,
    'FILTER: from=%s to=%s size=%d bytes=%d host-ip=%s host-name=%s helo=%s',
    $from, join(',', @recipients), $size, $nbytes, $host_ip, $host_name, $helo);

# Re-inject via sendmail
open(my $mail, '<', $tmp) or do {
    syslog(LOG_ERR, 'FILTER_ERROR: cannot open temp file: %s', $!);
    unlink($tmp);
    closelog();
    exit EX_TEMPFAIL;
};

my $pid = fork();
unless (defined $pid) {
    syslog(LOG_ERR, 'FILTER_ERROR: fork failed: %s', $!);
    close($mail);
    unlink($tmp);
    closelog();
    exit EX_TEMPFAIL;
}

if ($pid == 0) {
    open(STDIN, '<&', $mail) or _exit(EX_TEMPFAIL);
    close($mail);
    exec(SENDMAIL, '-i', '-f', $from, '--', @recipients)
        or do { print STDERR "filter: exec sendmail: $!\n"; POSIX::_exit(EX_TEMPFAIL); };
}

close($mail);
waitpid($pid, 0);
my $exit_code = $? >> 8;
unlink($tmp);

if ($? == -1 || $exit_code != 0) {
    syslog(LOG_ERR, 'FILTER_ERROR: sendmail failed (status=%d)', $exit_code);
    closelog();
    exit EX_TEMPFAIL;
}

syslog(LOG_INFO, 'FILTER_ACCEPT: from=%s recipients=%d', $from, scalar @recipients);
closelog();
exit EX_OK;
