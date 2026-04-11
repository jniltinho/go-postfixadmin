package transport

import "gorm.io/gorm"

// GetTransport resolves the transport for an email address or domain in a
// single query. When email is empty, only domain-level rules apply.
//
// Priority:
//  1. mailbox.transport        WHERE username = email  (per-user override)
//  2. transport_list.transport WHERE domain   = domain (explicit domain rule)
//  3. domain.transport         WHERE domain   = domain (default domain transport)
func GetTransport(db *gorm.DB, email, domain, hostname, localDelivery, delivery string) (string, error) {
	var result struct {
		DomainTransport        string `gorm:"column:domain_transport"`
		TransportListTransport string `gorm:"column:transport_list_transport"`
		MailboxTransport       string `gorm:"column:mailbox_transport"`
	}

	err := db.Raw(`
		SELECT
			d.transport  AS domain_transport,
			tl.transport AS transport_list_transport,
			m.transport  AS mailbox_transport
		FROM domain d
		LEFT JOIN transport_list tl ON tl.domain  = d.domain AND tl.active = true
		LEFT JOIN mailbox        m  ON m.username = ?        AND m.active  = true
		WHERE d.domain = ? AND d.active = true
		LIMIT 1
	`, email, domain).Scan(&result).Error
	if err != nil {
		return "", err
	}

	if result.MailboxTransport != "" && result.MailboxTransport != "virtual" {
		return formatTransport(result.MailboxTransport, hostname, localDelivery, delivery), nil
	}
	if result.TransportListTransport != "" {
		return formatTransport(result.TransportListTransport, hostname, localDelivery, delivery), nil
	}
	if result.DomainTransport != "" && result.DomainTransport != "virtual" {
		return formatTransport(result.DomainTransport, hostname, localDelivery, delivery), nil
	}

	return "", nil
}

func formatTransport(dest, hostname, localDelivery, delivery string) string {
	if hostname != "" && dest == "smtp:"+hostname {
		return "lmtp:unix:private/dovecot-lmtp"
	}
	if localDelivery != "" && dest == localDelivery {
		return delivery
	}
	return dest
}
