package kv

import "context"

type Iterator interface {
	// Ascend runs the user callback with key-value pair for a selected range in
	// ascending order.  Range can be selected with begin and end parameters.
	//
	// Iterated range is [begin, end) and ascending order requires begin <
	// end. User callback is not run when begin >= end, except when both are
	// empty strings.
	//
	// When both begin and end are empty strings, then user callback is run for
	// all key-value pairs in the database. Note that, if the database backend
	// allows empty string "" as a valid key, then this callback behavior is
	// different for empty keys and non-empty keys.
	//
	// Also, note that, when called from transactions, every key-value pair
	// passed to the callback is considered as READ, which may impact the
	// transaction Commit behavior.
	Ascend(ctx context.Context, begin, end string, cb func(context.Context, string, string) error) error

	// Descend is similar to Ascend, but iterates in reverse direction. Iterated
	// range is still [begin, end), but descending order requires begin >
	// end. Similar to the Ascend, user callback is not run when begin <= end,
	// except when both are empty string. Also, see the note if empty string is a
	// valid key.
	Descend(ctx context.Context, begin, end string, cb func(context.Context, string, string) error) error
}
