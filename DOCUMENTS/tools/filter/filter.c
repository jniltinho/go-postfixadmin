/*
 * filter.c — Postfix content filter (C version)
 *
 * Usage (configured in Postfix master.cf):
 *   filter unix - n n - 10 pipe
 *     flags=Rq user=filter argv=/usr/local/bin/filter
 *     --stress ${stress} --size ${size}
 *     --host-ip ${client_address} --host-name ${client_name}
 *     --helo ${client_helo} --from ${sender}
 *     -- ${recipient}
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <errno.h>
#include <time.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <sys/wait.h>
#include <syslog.h>
#include <fcntl.h>

/* Exit codes from <sysexits.h> */
#define EX_OK       0
#define EX_TEMPFAIL 75

#define INSPECT_DIR "/var/spool/filter"
#define SENDMAIL    "/usr/sbin/sendmail"

#define MAX_RECIPIENTS 256
#define MAX_ARG_LEN    512
#define BUF_SIZE       65536

static char  stress[MAX_ARG_LEN]    = "";
static long  size                    = 0;
static char  host_ip[MAX_ARG_LEN]   = "";
static char  host_name[MAX_ARG_LEN] = "";
static char  helo[MAX_ARG_LEN]      = "";
static char  from[MAX_ARG_LEN]      = "";

static char *recipients[MAX_RECIPIENTS];
static int   nrecipients = 0;

static void str_tolower(char *s) {
    for (; *s; s++)
        if (*s >= 'A' && *s <= 'Z') *s += 32;
}

static void usage(const char *prog) {
    fprintf(stdout,
        "Usage: %s [OPTIONS] -- RECIPIENT [RECIPIENT ...]\n"
        "\n"
        "Postfix content filter: saves the message to a temp file,\n"
        "logs it to syslog (mail.info) and re-injects via sendmail.\n"
        "\n"
        "Options:\n"
        "  --from      <addr>   Envelope sender address (lowercased)\n"
        "  --host-ip   <ip>     Client IP address\n"
        "  --host-name <name>   Client hostname\n"
        "  --helo      <name>   Client HELO/EHLO name\n"
        "  --size      <bytes>  Message size hint from Postfix\n"
        "  --stress    <value>  Postfix stress flag\n"
        "  --                   Start of recipient list\n"
        "  --help               Show this help and exit\n"
        "\n"
        "Paths:\n"
        "  Temp dir : %s\n"
        "  Sendmail : %s\n"
        "\n"
        "Exit codes:\n"
        "  0   Success\n"
        "  75  Temporary failure (Postfix will retry)\n"
        "\n"
        "Postfix master.cf configuration:\n"
        "\n"
        "  # Step 1: redirect smtp input through the filter\n"
        "  smtp      inet  n  -  y  -  -  smtpd\n"
        "    -o content_filter=filter:dummy\n"
        "\n"
        "  # Step 2: declare the filter service\n"
        "  filter    unix  -  n  n  -  10  pipe\n"
        "    flags=Rq user=filter null_sender=\n"
        "    argv=%s\n"
        "      --stress ${stress} --size ${size}\n"
        "      --host-ip ${client_address}\n"
        "      --host-name ${client_name}\n"
        "      --helo ${client_helo}\n"
        "      --from ${sender}\n"
        "      -- ${recipient}\n"
        "\n"
        "  # Step 3: re-injection bypass (avoid filter loop)\n"
        "  127.0.0.1:10025 inet  n  -  n  -  -  smtpd\n"
        "    -o content_filter=\n"
        "    -o receive_override_options=no_unknown_recipient_checks\n",
        prog, INSPECT_DIR, SENDMAIL, prog);
}

static void parse_args(int argc, char **argv) {
    int rcpt = 0;
    for (int i = 1; i < argc; i++) {
        if (rcpt) {
            char *r = argv[i];
            while (*r == ' ') r++;
            if (*r && nrecipients < MAX_RECIPIENTS)
                recipients[nrecipients++] = r;
            continue;
        }
        if (strcmp(argv[i], "--help") == 0) {
            usage(argv[0]);
            exit(EX_OK);
        } else if (strcmp(argv[i], "--") == 0) {
            rcpt = 1;
        } else if (strcmp(argv[i], "--stress") == 0 && i+1 < argc) {
            snprintf(stress, sizeof(stress), "%s", argv[++i]);
        } else if (strcmp(argv[i], "--size") == 0 && i+1 < argc) {
            size = atol(argv[++i]);
        } else if (strcmp(argv[i], "--host-ip") == 0 && i+1 < argc) {
            snprintf(host_ip, sizeof(host_ip), "%s", argv[++i]);
        } else if (strcmp(argv[i], "--host-name") == 0 && i+1 < argc) {
            snprintf(host_name, sizeof(host_name), "%s", argv[++i]);
        } else if (strcmp(argv[i], "--helo") == 0 && i+1 < argc) {
            snprintf(helo, sizeof(helo), "%s", argv[++i]);
        } else if (strcmp(argv[i], "--from") == 0 && i+1 < argc) {
            snprintf(from, sizeof(from), "%s", argv[++i]);
            str_tolower(from);
        }
    }
}

static int ensure_dir(const char *path) {
    struct stat st;
    if (stat(path, &st) == 0)
        return S_ISDIR(st.st_mode) ? 0 : -1;
    return mkdir(path, 0750);
}

static long copy_fd(int fd_in, int fd_out) {
    static char buf[BUF_SIZE];
    long total = 0;
    ssize_t n;
    while ((n = read(fd_in, buf, sizeof(buf))) > 0) {
        ssize_t w = 0;
        while (w < n) {
            ssize_t r = write(fd_out, buf + w, (size_t)(n - w));
            if (r < 0) return -1;
            w += r;
        }
        total += n;
    }
    return (n < 0) ? -1 : total;
}

static char *join_recipients(void) {
    size_t total = 0;
    for (int i = 0; i < nrecipients; i++)
        total += strlen(recipients[i]) + 1;
    if (total == 0) total = 1;
    char *buf = malloc(total);
    if (!buf) return NULL;
    buf[0] = '\0';
    for (int i = 0; i < nrecipients; i++) {
        if (i) strcat(buf, ",");
        strcat(buf, recipients[i]);
    }
    return buf;
}

int main(int argc, char **argv) {
    parse_args(argc, argv);

    openlog("postfix/filter", LOG_PID, LOG_MAIL);

    if (nrecipients == 0) {
        syslog(LOG_ERR, "FILTER_ERROR: no recipients");
        closelog();
        return EX_TEMPFAIL;
    }

    if (ensure_dir(INSPECT_DIR) != 0) {
        syslog(LOG_ERR, "FILTER_ERROR: cannot create %s: %s",
               INSPECT_DIR, strerror(errno));
        closelog();
        return EX_TEMPFAIL;
    }

    /* Temp file: /var/spool/filter/<timestamp>_<pid>.eml */
    char tmp_path[512];
    snprintf(tmp_path, sizeof(tmp_path), "%s/%ld_%d.eml",
             INSPECT_DIR, (long)time(NULL), (int)getpid());

    int fd = open(tmp_path, O_WRONLY | O_CREAT | O_EXCL, 0640);
    if (fd < 0) {
        syslog(LOG_ERR, "FILTER_ERROR: cannot create temp file %s: %s",
               tmp_path, strerror(errno));
        closelog();
        return EX_TEMPFAIL;
    }

    long nbytes = copy_fd(STDIN_FILENO, fd);
    close(fd);

    if (nbytes < 0) {
        syslog(LOG_ERR, "FILTER_ERROR: cannot save mail: %s", strerror(errno));
        unlink(tmp_path);
        closelog();
        return EX_TEMPFAIL;
    }

    char *rcpt_str = join_recipients();
    syslog(LOG_INFO,
           "FILTER: from=%s to=%s size=%ld bytes=%ld host-ip=%s host-name=%s helo=%s",
           from, rcpt_str ? rcpt_str : "", size, nbytes,
           host_ip, host_name, helo);
    free(rcpt_str);

    /* Build argv for sendmail: sendmail -i -f <from> -- <rcpt...> */
    /* slots: [sendmail, -i, -f, from, --, rcpt0..rcptN, NULL] */
    int sm_argc = 5 + nrecipients + 1;
    char **sm_argv = calloc((size_t)sm_argc, sizeof(char *));
    if (!sm_argv) {
        syslog(LOG_ERR, "FILTER_ERROR: out of memory");
        unlink(tmp_path);
        closelog();
        return EX_TEMPFAIL;
    }
    sm_argv[0] = SENDMAIL;
    sm_argv[1] = "-i";
    sm_argv[2] = "-f";
    sm_argv[3] = from;
    sm_argv[4] = "--";
    for (int i = 0; i < nrecipients; i++)
        sm_argv[5 + i] = recipients[i];
    sm_argv[5 + nrecipients] = NULL;

    int in_fd = open(tmp_path, O_RDONLY);
    if (in_fd < 0) {
        syslog(LOG_ERR, "FILTER_ERROR: cannot open temp file: %s", strerror(errno));
        free(sm_argv);
        unlink(tmp_path);
        closelog();
        return EX_TEMPFAIL;
    }

    pid_t pid = fork();
    if (pid < 0) {
        syslog(LOG_ERR, "FILTER_ERROR: fork failed: %s", strerror(errno));
        close(in_fd);
        free(sm_argv);
        unlink(tmp_path);
        closelog();
        return EX_TEMPFAIL;
    }

    if (pid == 0) {
        /* Child: redirect temp file to stdin and exec sendmail */
        dup2(in_fd, STDIN_FILENO);
        close(in_fd);
        execv(SENDMAIL, sm_argv);
        fprintf(stderr, "filter: execv %s: %s\n", SENDMAIL, strerror(errno));
        _exit(EX_TEMPFAIL);
    }

    close(in_fd);
    free(sm_argv);

    int status;
    waitpid(pid, &status, 0);
    unlink(tmp_path);

    if (!WIFEXITED(status) || WEXITSTATUS(status) != 0) {
        syslog(LOG_ERR, "FILTER_ERROR: sendmail failed (status=%d)", status);
        closelog();
        return EX_TEMPFAIL;
    }

    syslog(LOG_INFO, "FILTER_ACCEPT: from=%s recipients=%d", from, nrecipients);
    closelog();
    return EX_OK;
}
