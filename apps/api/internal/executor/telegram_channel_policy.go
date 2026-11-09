package executor

import "encoding/json"

// shouldSkipTelegramBotReplyToChannel blocks any relay bot reply in Telegram channels.
func shouldSkipTelegramBotReplyToChannel(runInput []byte) bool {
	var root map[string]any
	if len(runInput) == 0 || json.Unmarshal(runInput, &root) != nil {
		return false
	}
	if _, ok := root["telegram_relay"]; !ok {
		return false
	}
	chatType := stringAtRunPath(root, "telegram_relay.chat.type")
	updateType := stringAtRunPath(root, "telegram_relay.update_type")
	return chatType == "channel" ||
		updateType == "channel_post" ||
		updateType == "edited_channel_post"
}
