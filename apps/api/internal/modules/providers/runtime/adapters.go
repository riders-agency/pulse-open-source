package runtime

var providerActionAdapters = []providerActionAdapter{
	{typeKey: "ai.apinet.normalize_text", execute: executeAPINETNormalizeTextAdapter},
	{typeKey: "telegram.send_message", execute: executeTelegramSendMessageAdapter},
	{typeKey: "telegram.edit_message_text", execute: executeTelegramEditMessageTextAdapter},
	{typeKey: "telegram.answer_callback_query", execute: executeTelegramAnswerCallbackQueryAdapter},
	{typeKey: "telegram.pin_message", execute: executeTelegramPinMessageAdapter},
	{typeKey: "telegram.unpin_message", execute: executeTelegramUnpinMessageAdapter},
	{typeKey: "telegram.send_document", execute: executeTelegramSendDocumentAdapter},
	{typeKey: "max.send_message", execute: executeMaxSendMessageAdapter},
	{typeKey: "email.send", execute: executeEmailSendAdapter},
	{typeKey: "uon.create_lead", execute: executeUONCreateLeadAdapter},
	{typeKey: "yclients.certificate.issue", execute: executeYClientsCertificateAdapter},
}
