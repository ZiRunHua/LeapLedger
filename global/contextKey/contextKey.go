package contextKey

type ContextKey string

const Tx ContextKey = "Tx"

type GinParamKey string

const AccountId GinParamKey = "accountId"
const TransactionTimingId GinParamKey = "timingId"
