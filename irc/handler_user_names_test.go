package irc

import (
	"testing"
)

func TestUserHandlerNames(t *testing.T) {
	state := make(chan State, 1)
	handler := func() Handler { return NewUserHandler(state, "nick") }
	testHandler(t, "UserHandler-NAMES", state, handler, []handlerTest{
		{
			desc: "names successful",
			in:   []Message{CmdNames.WithParams("#channel")},
			wantNicks: map[string]mockConnection{
				"nick": mockConnection{
					messages: []Message{
						ReplyNamReply,
						ReplyNamReply,
						ReplyNamReply,
						ReplyNamReply,
						ReplyEndOfNames,
					},
				},
			},
			state: newMockState().
				withChannel("#channel", "", "").
				withUser("nick", "#channel").
				withUser("foo", "#channel").
				withUser("bar", "#channel").
				// User baz should be listed even though they are invisible because
				// they share a channel with the user that is requesting names.
				withUser("baz", "#channel").
				withUserMode("baz", "i"),
		},
		{
			desc: "names all",
			in:   []Message{CmdNames.WithParams()},
			wantNicks: map[string]mockConnection{
				"nick": mockConnection{
					messages: []Message{
						ReplyNamReply,
						ReplyEndOfNames,
						ReplyNamReply,
						ReplyEndOfNames,
					},
				},
			},
			state: newMockState().
				withChannel("#foo", "", "").
				withChannel("#bar", "", "").
				withUser("nick").
				withUser("foo", "#foo").
				withUser("bar", "#bar").
				// User baz should not be listed because their are invisible.
				withUser("baz", "#bar").
				withUserMode("baz", "i"),
		},
		{
			desc: "names all secret",
			in:   []Message{CmdNames.WithParams()},
			wantNicks: map[string]mockConnection{
				"nick": mockConnection{
					messages: []Message{},
				},
			},
			state: newMockState().
				withChannel("#foo", "s", "").
				withUser("nick").
				withUser("foo", "#foo"),
		},
		{
			desc: "names successful private",
			in:   []Message{CmdNames.WithParams("#channel")},
			wantNicks: map[string]mockConnection{
				"nick": mockConnection{
					messages: []Message{
						ReplyNamReply,
						ReplyNamReply,
						ReplyNamReply,
						ReplyEndOfNames,
					},
				},
			},
			state: newMockState().
				withChannel("#channel", "p", "").
				withUser("nick", "#channel").
				withUser("foo", "#channel").
				withUser("bar", "#channel"),
		},
		{
			desc: "names fails private",
			in:   []Message{CmdNames.WithParams("#channel")},
			wantNicks: map[string]mockConnection{
				"nick": mockConnection{
					messages: []Message{},
				},
			},
			state: newMockState().
				withChannel("#channel", "p", "").
				withUser("nick").
				withUser("foo", "#channel").
				withUser("bar", "#channel"),
		},
	})
}
