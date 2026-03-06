package smtpsender

type Option func(*SmtpSender)

func Host(host string) Option {
	return func(ms *SmtpSender) {
		ms.host = host
	}
}

func Port(port string) Option {
	return func(ms *SmtpSender) {
		ms.port = port
	}
}

func Username(u string) Option {
	return func(ms *SmtpSender) {
		ms.username = u
	}
}

func Password(p string) Option {
	return func(ms *SmtpSender) {
		ms.password = p
	}
}
