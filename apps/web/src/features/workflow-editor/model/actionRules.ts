import type { ActionKind } from "./types";

const startActionTypes: readonly ActionKind[] = [
  "mailbox.incoming.trigger",
  "webhook.trigger",
  "telegram.relay.trigger",
  "schedule.trigger",
  "manual.trigger",
  "clients.manual_payment.trigger",
];

const retryExcludedActionTypes: readonly ActionKind[] = [
  ...startActionTypes,
  "clients.payment_due.trigger",
  "clients.payment_cycle.collect",
  "bot.set_state",
];

export function startsRunWithoutIncomingEdge(type: ActionKind) {
  return startActionTypes.includes(type);
}

export function supportsRetryPolicy(type: ActionKind) {
  return !retryExcludedActionTypes.includes(type);
}
