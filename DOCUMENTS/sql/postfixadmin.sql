

CREATE TABLE `admin` (
  `username` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `password` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `created` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `modified` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `superadmin` tinyint(1) NOT NULL DEFAULT '0',
  `phone` varchar(30) CHARACTER SET utf8 NOT NULL DEFAULT '',
  `email_other` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '',
  `token` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '',
  `token_validity` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `totp_secret` varchar(255) CHARACTER SET utf8 DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci COMMENT='Postfix Admin - Virtual Admins';


CREATE TABLE `alias` (
  `address` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `goto` text COLLATE utf8_general_ci NOT NULL,
  `domain` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `created` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `modified` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `active` tinyint(1) NOT NULL DEFAULT '1'
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci COMMENT='Postfix Admin - Virtual Aliases';


CREATE TABLE `alias_domain` (
  `alias_domain` varchar(255) COLLATE utf8_general_ci NOT NULL DEFAULT '',
  `target_domain` varchar(255) COLLATE utf8_general_ci NOT NULL DEFAULT '',
  `created` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `modified` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `active` tinyint(1) NOT NULL DEFAULT '1'
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci COMMENT='Postfix Admin - Domain Aliases';


CREATE TABLE `config` (
  `id` int(11) NOT NULL,
  `name` varchar(20) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '',
  `value` varchar(20) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='PostfixAdmin settings';


CREATE TABLE `dkim` (
  `id` int(11) NOT NULL,
  `domain_name` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `description` varchar(255) COLLATE utf8_general_ci DEFAULT '',
  `selector` varchar(63) COLLATE utf8_general_ci NOT NULL DEFAULT 'default',
  `private_key` text COLLATE utf8_general_ci,
  `public_key` text COLLATE utf8_general_ci,
  `created` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `modified` datetime NOT NULL DEFAULT '2000-01-01 00:00:00'
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci COMMENT='Postfix Admin - OpenDKIM Key Table';


CREATE TABLE `dkim_signing` (
  `id` int(11) NOT NULL,
  `author` varchar(255) COLLATE utf8_general_ci NOT NULL DEFAULT '',
  `dkim_id` int(11) NOT NULL,
  `created` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `modified` datetime NOT NULL DEFAULT '2000-01-01 00:00:00'
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci COMMENT='Postfix Admin - OpenDKIM Signing Table';


CREATE TABLE `domain` (
  `domain` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `description` varchar(255) CHARACTER SET utf8 NOT NULL,
  `aliases` int(10) NOT NULL DEFAULT '0',
  `mailboxes` int(10) NOT NULL DEFAULT '0',
  `maxquota` bigint(20) NOT NULL DEFAULT '0',
  `quota` bigint(20) NOT NULL DEFAULT '0',
  `transport` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `backupmx` tinyint(1) NOT NULL DEFAULT '0',
  `created` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `modified` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `password_expiry` int(11) DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci COMMENT='Postfix Admin - Virtual Domains';


CREATE TABLE `domain_admins` (
  `username` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `domain` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `created` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci COMMENT='Postfix Admin - Domain Admins';


CREATE TABLE `fetchmail` (
  `id` int(11) UNSIGNED NOT NULL,
  `mailbox` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `src_server` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `src_auth` enum('password','kerberos_v5','kerberos','kerberos_v4','gssapi','cram-md5','otp','ntlm','msn','ssh','any') COLLATE utf8_general_ci DEFAULT NULL,
  `src_user` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `src_password` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `src_folder` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `poll_time` int(11) UNSIGNED NOT NULL DEFAULT '10',
  `fetchall` tinyint(1) UNSIGNED NOT NULL DEFAULT '0',
  `keep` tinyint(1) UNSIGNED NOT NULL DEFAULT '0',
  `protocol` enum('POP3','IMAP','POP2','ETRN','AUTO') COLLATE utf8_general_ci DEFAULT NULL,
  `usessl` tinyint(1) UNSIGNED NOT NULL DEFAULT '0',
  `extra_options` text COLLATE utf8_general_ci,
  `returned_text` text COLLATE utf8_general_ci,
  `mda` varchar(255) COLLATE utf8_general_ci DEFAULT NULL,
  `date` timestamp NOT NULL DEFAULT '2000-01-01 03:00:00',
  `sslcertck` tinyint(1) NOT NULL DEFAULT '0',
  `sslcertpath` varchar(255) CHARACTER SET utf8 DEFAULT '',
  `sslfingerprint` varchar(255) COLLATE utf8_general_ci DEFAULT '',
  `domain` varchar(255) COLLATE utf8_general_ci DEFAULT '',
  `active` tinyint(1) NOT NULL DEFAULT '0',
  `created` timestamp NOT NULL DEFAULT '2000-01-01 03:00:00',
  `modified` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `src_port` int(11) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;


CREATE TABLE `log` (
  `timestamp` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `username` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `domain` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `action` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `data` text COLLATE utf8_general_ci NOT NULL,
  `id` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci COMMENT='Postfix Admin - Log';


CREATE TABLE `mailbox` (
  `username` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `password` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `name` varchar(255) CHARACTER SET utf8 NOT NULL,
  `maildir` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `quota` bigint(20) NOT NULL DEFAULT '0',
  `local_part` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `domain` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `created` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `modified` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `phone` varchar(30) CHARACTER SET utf8 NOT NULL DEFAULT '',
  `email_other` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '',
  `token` varchar(255) CHARACTER SET utf8 NOT NULL DEFAULT '',
  `token_validity` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `password_expiry` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `totp_secret` varchar(255) CHARACTER SET utf8 DEFAULT NULL,
  `smtp_active` tinyint(1) NOT NULL DEFAULT '1'
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci COMMENT='Postfix Admin - Virtual Mailboxes';


CREATE TABLE `mailbox_app_password` (
  `id` int(11) NOT NULL,
  `username` varchar(255) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  `password_hash` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE `quota` (
  `username` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `path` varchar(100) COLLATE utf8_general_ci NOT NULL,
  `current` bigint(20) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;


CREATE TABLE `quota2` (
  `username` varchar(100) COLLATE utf8_general_ci NOT NULL,
  `bytes` bigint(20) NOT NULL DEFAULT '0',
  `messages` int(11) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;


CREATE TABLE `totp_exception_address` (
  `id` int(11) NOT NULL,
  `ip` varchar(46) NOT NULL,
  `username` varchar(255) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE `vacation` (
  `email` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `subject` varchar(255) CHARACTER SET utf8 NOT NULL,
  `body` text CHARACTER SET utf8 NOT NULL,
  `cache` text COLLATE utf8_general_ci NOT NULL,
  `domain` varchar(255) COLLATE utf8_general_ci NOT NULL,
  `created` datetime NOT NULL DEFAULT '2000-01-01 00:00:00',
  `active` tinyint(1) NOT NULL DEFAULT '1',
  `modified` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `activefrom` timestamp NOT NULL DEFAULT '2000-01-01 03:00:00',
  `activeuntil` timestamp NOT NULL DEFAULT '2038-01-18 03:00:00',
  `interval_time` int(11) NOT NULL DEFAULT '0'
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci COMMENT='Postfix Admin - Virtual Vacation';


CREATE TABLE `vacation_notification` (
  `on_vacation` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `notified` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '',
  `notified_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='Postfix Admin - Virtual Vacation Notifications';

--
-- Índices para tabelas despejadas
--

--
-- Índices de tabela `admin`
--
ALTER TABLE `admin`
  ADD PRIMARY KEY (`username`);

--
-- Índices de tabela `alias`
--
ALTER TABLE `alias`
  ADD PRIMARY KEY (`address`),
  ADD KEY `domain` (`domain`);

--
-- Índices de tabela `alias_domain`
--
ALTER TABLE `alias_domain`
  ADD PRIMARY KEY (`alias_domain`),
  ADD KEY `active` (`active`),
  ADD KEY `target_domain` (`target_domain`);

--
-- Índices de tabela `config`
--
ALTER TABLE `config`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `name` (`name`);

--
-- Índices de tabela `dkim`
--
ALTER TABLE `dkim`
  ADD PRIMARY KEY (`id`),
  ADD KEY `domain_name` (`domain_name`,`description`);

--
-- Índices de tabela `dkim_signing`
--
ALTER TABLE `dkim_signing`
  ADD PRIMARY KEY (`id`),
  ADD KEY `author` (`author`),
  ADD KEY `dkim_id` (`dkim_id`);

--
-- Índices de tabela `domain`
--
ALTER TABLE `domain`
  ADD PRIMARY KEY (`domain`);

--
-- Índices de tabela `domain_admins`
--
ALTER TABLE `domain_admins`
  ADD PRIMARY KEY (`id`),
  ADD KEY `username` (`username`);

--
-- Índices de tabela `fetchmail`
--
ALTER TABLE `fetchmail`
  ADD PRIMARY KEY (`id`);

--
-- Índices de tabela `log`
--
ALTER TABLE `log`
  ADD PRIMARY KEY (`id`),
  ADD KEY `timestamp` (`timestamp`),
  ADD KEY `domain_timestamp` (`domain`,`timestamp`);

--
-- Índices de tabela `mailbox`
--
ALTER TABLE `mailbox`
  ADD PRIMARY KEY (`username`),
  ADD KEY `domain` (`domain`);

--
-- Índices de tabela `mailbox_app_password`
--
ALTER TABLE `mailbox_app_password`
  ADD PRIMARY KEY (`id`);

--
-- Índices de tabela `quota`
--
ALTER TABLE `quota`
  ADD PRIMARY KEY (`username`,`path`);

--
-- Índices de tabela `quota2`
--
ALTER TABLE `quota2`
  ADD PRIMARY KEY (`username`);

--
-- Índices de tabela `totp_exception_address`
--
ALTER TABLE `totp_exception_address`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `ip_user` (`ip`,`username`);

--
-- Índices de tabela `vacation`
--
ALTER TABLE `vacation`
  ADD PRIMARY KEY (`email`),
  ADD KEY `email` (`email`);

--
-- Índices de tabela `vacation_notification`
--
ALTER TABLE `vacation_notification`
  ADD PRIMARY KEY (`on_vacation`,`notified`);

--
-- AUTO_INCREMENT para tabelas despejadas
--

--
-- AUTO_INCREMENT de tabela `config`
--
ALTER TABLE `config`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT de tabela `dkim`
--
ALTER TABLE `dkim`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT de tabela `dkim_signing`
--
ALTER TABLE `dkim_signing`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT de tabela `domain_admins`
--
ALTER TABLE `domain_admins`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT de tabela `fetchmail`
--
ALTER TABLE `fetchmail`
  MODIFY `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT de tabela `log`
--
ALTER TABLE `log`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT de tabela `mailbox_app_password`
--
ALTER TABLE `mailbox_app_password`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- AUTO_INCREMENT de tabela `totp_exception_address`
--
ALTER TABLE `totp_exception_address`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;

--
-- Restrições para tabelas despejadas
--

--
-- Restrições para tabelas `dkim`
--
ALTER TABLE `dkim`
  ADD CONSTRAINT `dkim_ibfk_1` FOREIGN KEY (`domain_name`) REFERENCES `domain` (`domain`) ON DELETE CASCADE;

--
-- Restrições para tabelas `dkim_signing`
--
ALTER TABLE `dkim_signing`
  ADD CONSTRAINT `dkim_signing_ibfk_1` FOREIGN KEY (`dkim_id`) REFERENCES `dkim` (`id`) ON DELETE CASCADE;

--
-- Restrições para tabelas `vacation_notification`
--
ALTER TABLE `vacation_notification`
  ADD CONSTRAINT `vacation_notification_pkey` FOREIGN KEY (`on_vacation`) REFERENCES `vacation` (`email`) ON DELETE CASCADE;
COMMIT;
