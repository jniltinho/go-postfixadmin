#!/bin/bash

# This script receives two parameters passed by dovecot:
# 1 = Usage percentage (e.g., 80, 95)
# 2 = Logged in user's email address

PERCENT=$1
USER=$2

cat << EOF | /usr/sbin/sendmail -f "postmaster@$(hostname -d)" -t "$USER"
From: postmaster@$(hostname -d)
To: $USER
Subject: Email Quota Warning (${PERCENT}%)
Content-Type: text/plain; charset="utf-8"

Dear user,

Your mailbox has reached ${PERCENT}% of its total storage capacity.
Please delete old or unwanted messages to free up space and avoid being blocked from receiving new messages.

Best regards,
System Administrator
EOF
