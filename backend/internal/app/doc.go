// Package app is the composition root for the Inkless CMS API process.
//
// cmd/server (and any future entrypoints) should only load configuration and
// call [New] + [App.Run]. Dependency wiring, HTTP registration, and background
// workers live here so startup behavior stays in one place.
package app
