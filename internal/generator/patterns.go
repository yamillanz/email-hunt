package generator

type pattern struct {
	name string
	fn   func(parts nameParts) string
}

func buildPatterns() []pattern {
	return []pattern{
		{name: "{first}", fn: func(parts nameParts) string { return parts.first }},
		{name: "{last}", fn: func(parts nameParts) string { return parts.last }},
		{name: "{first}{last}", fn: func(parts nameParts) string { return parts.first + parts.last }},
		{name: "{first}.{last}", fn: func(parts nameParts) string { return parts.first + "." + parts.last }},
		{name: "{first}_{last}", fn: func(parts nameParts) string { return parts.first + "_" + parts.last }},
		{name: "{first}-{last}", fn: func(parts nameParts) string { return parts.first + "-" + parts.last }},
		{name: "{f}{last}", fn: func(parts nameParts) string {
			f, _ := parts.initials()
			return f + parts.last
		}},
		{name: "{f}.{last}", fn: func(parts nameParts) string {
			f, _ := parts.initials()
			return f + "." + parts.last
		}},
		{name: "{first}{l}", fn: func(parts nameParts) string {
			_, l := parts.initials()
			return parts.first + l
		}},
		{name: "{first}.{l}", fn: func(parts nameParts) string {
			_, l := parts.initials()
			return parts.first + "." + l
		}},
		{name: "{f}{l}", fn: func(parts nameParts) string {
			f, l := parts.initials()
			return f + l
		}},
		{name: "{f}.{l}", fn: func(parts nameParts) string {
			f, l := parts.initials()
			return f + "." + l
		}},
		{name: "{last}{first}", fn: func(parts nameParts) string { return parts.last + parts.first }},
		{name: "{last}.{first}", fn: func(parts nameParts) string { return parts.last + "." + parts.first }},
		{name: "{last}{f}", fn: func(parts nameParts) string {
			f, _ := parts.initials()
			return parts.last + f
		}},
		{name: "{last}.{f}", fn: func(parts nameParts) string {
			f, _ := parts.initials()
			return parts.last + "." + f
		}},
		{name: "{l}{first}", fn: func(parts nameParts) string {
			_, l := parts.initials()
			return l + parts.first
		}},
		{name: "{l}.{first}", fn: func(parts nameParts) string {
			_, l := parts.initials()
			return l + "." + parts.first
		}},
		{name: "{first}{m}{last}", fn: func(parts nameParts) string {
			return parts.first + parts.middleInitials() + parts.last
		}},
		{name: "{first}.{m}.{last}", fn: func(parts nameParts) string {
			return parts.first + "." + parts.middleInitials() + "." + parts.last
		}},
		{name: "{f}{m}{last}", fn: func(parts nameParts) string {
			f, _ := parts.initials()
			return f + parts.middleInitials() + parts.last
		}},
		{name: "{first}{middle}{last}", fn: func(parts nameParts) string {
			return parts.first + parts.middleJoined("") + parts.last
		}},
		{name: "{first}.{middle}.{last}", fn: func(parts nameParts) string {
			return parts.first + "." + parts.middleJoined(".") + "." + parts.last
		}},
		{name: "{first}-{middle}-{last}", fn: func(parts nameParts) string {
			return parts.first + "-" + parts.middleJoined("-") + "-" + parts.last
		}},
		{name: "{f}{m}{l}", fn: func(parts nameParts) string {
			f, l := parts.initials()
			return f + parts.middleInitials() + l
		}},
	}
}
