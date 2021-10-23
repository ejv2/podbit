// Package components implements re-usable sections of UI which can be
// used by multiple sections of the program in multiple different ways.
//
// They are designed to be as self-contained and as simple to use as
// possible, with immediate-mode UI used as a reference model.
//
// As a general rule, it is safe to assume that none of this package is
// thread-safe, as it is only really designed for use by the UI thread.
// That being said, there is nothing stopping you from connecting some
// methods up to channel listeners or just locking the UI using a mutex
// either.
package components
