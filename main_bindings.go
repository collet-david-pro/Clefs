//go:build bindings

package main

// Compilé uniquement par la génération des liaisons JavaScript (tag
// « bindings ») : main n'ouvre alors ni base ni dossiers de travail.
func init() { bindingsOnly = true }
