package cusCtx

type ContextKey string

const Claims ContextKey = "claims"

const UserId ContextKey = "userID"
const User ContextKey = "user"

const Account ContextKey = "account"
const AccountUser ContextKey = "accountUser"
const AccountId ContextKey = "accountID"

const Db ContextKey = "Db"
const Tx ContextKey = "Tx"
const TxCommit ContextKey = "TxCommit"
