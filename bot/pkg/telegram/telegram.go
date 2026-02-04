package telegram

import (
	"context"
	"fmt"
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Message represents an incoming telegram message
type Message struct {
	ChatID    int64
	MessageID int
	UserID    int64
	Username  string
	Text      string
}

// Callback represents an inline button callback
type Callback struct {
	ID        string
	ChatID    int64
	MessageID int
	UserID    int64
	Username  string
	Data      string
}

// Button represents an inline keyboard button
type Button struct {
	Text string
	Data string
}

// Command represents a bot menu command
type Command struct {
	Command     string
	Description string
}

// Handler is called when a message is received
type Handler func(msg Message)

// CallbackHandler is called when an inline button is pressed
type CallbackHandler func(cb Callback)

// Bot wraps telegram bot API
type Bot struct {
	api             *tgbotapi.BotAPI
	handler         Handler
	callbackHandler CallbackHandler
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

// New creates a new telegram bot
func New(token string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Bot{
		api:    api,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// SetHandler sets the message handler
func (b *Bot) SetHandler(h Handler) {
	b.handler = h
}

// SetCallbackHandler sets the inline button callback handler
func (b *Bot) SetCallbackHandler(h CallbackHandler) {
	b.callbackHandler = h
}

// SetCommands sets the bot menu commands
func (b *Bot) SetCommands(commands []Command) error {
	cmds := make([]tgbotapi.BotCommand, len(commands))
	for i, c := range commands {
		cmds[i] = tgbotapi.BotCommand{
			Command:     c.Command,
			Description: c.Description,
		}
	}
	cfg := tgbotapi.NewSetMyCommands(cmds...)
	_, err := b.api.Request(cfg)
	if err != nil {
		return fmt.Errorf("set commands: %w", err)
	}
	return nil
}

// SetKeyboard sets the persistent reply keyboard
func (b *Bot) SetKeyboard(chatID int64, text string, buttons [][]string) error {
	rows := make([][]tgbotapi.KeyboardButton, len(buttons))
	for i, row := range buttons {
		rows[i] = make([]tgbotapi.KeyboardButton, len(row))
		for j, btn := range row {
			rows[i][j] = tgbotapi.NewKeyboardButton(btn)
		}
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{
		Keyboard:       rows,
		ResizeKeyboard: true,
	}
	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("set keyboard: %w", err)
	}
	return nil
}

// RemoveKeyboard removes the reply keyboard
func (b *Bot) RemoveKeyboard(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(false)
	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("remove keyboard: %w", err)
	}
	return nil
}

// Start begins listening for messages
func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		for {
			select {
			case <-b.ctx.Done():
				return
			case update := <-updates:
				// Handle callback query (inline button press)
				if update.CallbackQuery != nil {
					if b.callbackHandler != nil {
						cb := Callback{
							ID:        update.CallbackQuery.ID,
							ChatID:    update.CallbackQuery.Message.Chat.ID,
							MessageID: update.CallbackQuery.Message.MessageID,
							UserID:    update.CallbackQuery.From.ID,
							Username:  update.CallbackQuery.From.UserName,
							Data:      update.CallbackQuery.Data,
						}
						b.callbackHandler(cb)
					}
					continue
				}

				// Handle message
				if update.Message == nil {
					continue
				}
				if b.handler != nil {
					msg := Message{
						ChatID:    update.Message.Chat.ID,
						MessageID: update.Message.MessageID,
						UserID:    update.Message.From.ID,
						Username:  update.Message.From.UserName,
						Text:      update.Message.Text,
					}
					b.handler(msg)
				}
			}
		}
	}()

	log.Printf("Telegram bot started: @%s", b.api.Self.UserName)
	return nil
}

// Stop stops the bot
func (b *Bot) Stop() {
	b.cancel()
	b.api.StopReceivingUpdates()
	b.wg.Wait()
	log.Println("Telegram bot stopped")
}

// Send sends a text message to a chat
func (b *Bot) Send(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	return nil
}

// SendWithButtons sends a message with inline buttons
func (b *Bot) SendWithButtons(chatID int64, text string, buttons [][]Button) error {
	rows := make([][]tgbotapi.InlineKeyboardButton, len(buttons))
	for i, row := range buttons {
		rows[i] = make([]tgbotapi.InlineKeyboardButton, len(row))
		for j, btn := range row {
			rows[i][j] = tgbotapi.NewInlineKeyboardButtonData(btn.Text, btn.Data)
		}
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("send with buttons: %w", err)
	}
	return nil
}

// EditButtons edits the inline buttons of an existing message
func (b *Bot) EditButtons(chatID int64, messageID int, buttons [][]Button) error {
	rows := make([][]tgbotapi.InlineKeyboardButton, len(buttons))
	for i, row := range buttons {
		rows[i] = make([]tgbotapi.InlineKeyboardButton, len(row))
		for j, btn := range row {
			rows[i][j] = tgbotapi.NewInlineKeyboardButtonData(btn.Text, btn.Data)
		}
	}

	markup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	edit := tgbotapi.NewEditMessageReplyMarkup(chatID, messageID, markup)
	_, err := b.api.Send(edit)
	if err != nil {
		return fmt.Errorf("edit buttons: %w", err)
	}
	return nil
}

// EditText edits the text of an existing message
func (b *Bot) EditText(chatID int64, messageID int, text string) error {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	_, err := b.api.Send(edit)
	if err != nil {
		return fmt.Errorf("edit text: %w", err)
	}
	return nil
}

// AnswerCallback answers a callback query (removes loading state)
func (b *Bot) AnswerCallback(callbackID string, text string) error {
	callback := tgbotapi.NewCallback(callbackID, text)
	_, err := b.api.Request(callback)
	if err != nil {
		return fmt.Errorf("answer callback: %w", err)
	}
	return nil
}

// SendMarkdown sends a markdown formatted message
func (b *Bot) SendMarkdown(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("send markdown: %w", err)
	}
	return nil
}

// Reply sends a reply to a specific message
func (b *Bot) Reply(chatID int64, replyToID int, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = replyToID
	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("reply message: %w", err)
	}
	return nil
}

// Username returns the bot's username
func (b *Bot) Username() string {
	return b.api.Self.UserName
}
