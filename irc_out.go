package main

import (
	"fmt"
	"strings"
	"time"
)

type IrcNIH struct {
	Nick  string
	Ident string
	Host  string
}

func (nih IrcNIH) String() string {
	return fmt.Sprintf("%s!%s@%s", nih.Nick, nih.Ident, nih.Host)
}

type IrcMessage interface {
	ToIrc(ircd *Ircd) string
}

type IrcWelcomeBanner struct {
	Nick, Ident, Host string
}

func (msg IrcWelcomeBanner) ToIrc(ircd *Ircd) string {
	return fmt.Sprintf(":%s 001 %s :Welcome to the %s Internet Relay Chat network %s!%s@%s", ircd.node.Me.Name, msg.Nick, ircd.node.NetworkName(), msg.Nick, msg.Ident, msg.Host)
}

type IrcWelcomeHost struct {
	Nick string
}

func (msg IrcWelcomeHost) ToIrc(ircd *Ircd) string {
	return fmt.Sprintf(":%s 002 %s :Your host is %s, running version gossamer-dev", ircd.node.Me.Name, msg.Nick, ircd.node.Me.Name)
}

type IrcWelcomeCreated struct {
	Nick string
}

func (msg IrcWelcomeCreated) ToIrc(ircd *Ircd) string {
	return fmt.Sprintf(":%s 003 %s :This server was created %s", ircd.node.Me.Name, msg.Nick, ircd.time.Format(time.RFC1123))
}

type IrcWelcomeSupportedModes struct {
	Nick string
}

func (msg IrcWelcomeSupportedModes) ToIrc(ircd *Ircd) string {
	// TODO: Use real modes when we actually support some.
	return fmt.Sprintf(":%s 004 %s %s gossamer-dev CDFGNRSUWXabcdfgijklnopqrsuwxyz BIMNORSabcehiklmnopqstvz Iabehkloqv", ircd.node.Me.Name, msg.Nick, ircd.node.Me.Name)
}

type IrcWelcomeSupportedFeatures struct {
	Nick    string
	Feature map[string]string
}

func (msg IrcWelcomeSupportedFeatures) ToIrc(ircd *Ircd) string {
	list := make([]string, 0, len(msg.Feature))
	for k, v := range msg.Feature {
		if v != "" {
			list = append(list, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
		} else {
			list = append(list, strings.ToUpper(k))
		}
	}
	return fmt.Sprintf(":%s 005 %s %s :are supported by this server", ircd.node.Me.Name, msg.Nick, strings.Join(list, " "))
}

type IrcPrivateMessage struct {
	From    IrcNIH
	To      string
	Message string
}

func (msg IrcPrivateMessage) ToIrc(ircd *Ircd) string {
	return fmt.Sprintf(":%s PRIVMSG %s :%s", msg.From, msg.To, msg.Message)
}

type IrcNickInUse struct {
	Nick string
}

func (msg IrcNickInUse) ToIrc(ircd *Ircd) string {
	return fmt.Sprintf(":%s 433 %s :Nickname is already in use", ircd.node.Me.Name, msg.Nick)
}
