package cusCtx

type ContextKey string

const Db ContextKey = "Db"
const Tx ContextKey = "Tx"
const TxCommit ContextKey = "TxCommit"

type GinParamKey string

const AccountId GinParamKey = "accountId"
const TransactionTimingId GinParamKey = "timingId"
