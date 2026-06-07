package main

// Noun is the German singular/plural form of a resource name.
type Noun struct {
	Singular string
	Plural   string
}

// Translations provides German descriptions for commands.
//
// CRUD commands are built formulaically from the resource noun (Nouns, per tag)
// and a verb template. Non-CRUD actions get fixed text via Actions (key:
// "<group>/<name>"). If no translation exists, the English summary is used as a
// fallback.
type Translations struct {
	Nouns   map[string]Noun
	Actions map[string]string
}

func (t *Translations) short(c *Command) string {
	if s, ok := t.Actions[c.Group+"/"+c.Name]; ok {
		return s
	}
	if n, ok := t.Nouns[c.Tag]; ok {
		switch c.Name {
		case "list":
			return n.Plural + " auflisten"
		case "get":
			return n.Singular + " abrufen"
		case "create":
			return n.Singular + " erstellen"
		case "update":
			return n.Singular + " aktualisieren"
		case "patch":
			return n.Singular + " teilweise aktualisieren"
		case "delete":
			return n.Singular + " löschen"
		}
	}
	return c.Summary
}

// LoadTranslations returns the maintained translation table.
func LoadTranslations() *Translations {
	return &Translations{Nouns: nouns, Actions: actions}
}
