package web

import _ "embed"

// LinkInitializeJS is the window.Plaid shim served at the Plaid CDN path.
//
//go:embed link-initialize.js
var LinkInitializeJS []byte

// LinkHTML is the mock Link UI (bank/account dropdown) served at /link.
//
//go:embed link.html
var LinkHTML []byte
