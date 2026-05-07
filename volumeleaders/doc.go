// Package volumeleaders provides a small HTTP client for the VolumeLeaders web
// application endpoints used by browser-backed sessions.
//
// Callers supply authenticated browser cookies and the ASP.NET XSRF token in a
// Session, usually through NewSession or SessionFromCookies. The optional
// browserauth subpackage can discover local browser sessions for desktop
// automation without adding browser-store dependencies to this core package.
//
// The client sends browser-like XHR requests, encodes server-side DataTables
// form bodies, and decodes typed endpoint responses.
package volumeleaders
