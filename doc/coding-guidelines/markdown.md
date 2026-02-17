# Markdown Coding Guidelines

## General Rules

1. Do not use word decorations such as `**word**` or `*word*`.
2. Use hyphens or numbers for bullet points.
3. Keep sentences simple and concise.

## Heading Format Rules

### 1. Empty Lines Around Headings

Always include exactly one empty line before and after headings. Multiple empty lines should be avoided.

#### Good Example

```markdown
This is some text.

## Heading

This is the next paragraph.
```

#### Bad Example

```markdown
This is some text.
## Heading
This is the next paragraph.
```

Or

```markdown
This is some text.

## Heading


This is the next paragraph.
```

### 2. Space Between Hash and Heading Text

Include exactly one space between the hash symbols (`#`) and the heading text.

#### Good Example

```markdown
# Heading 1
## Heading 2
### Heading 3
```

#### Bad Example

```markdown
#Heading 1
##  Heading 2 (two spaces)
###Heading 3
```

### 3. One Sentence per List Item

Each list item should contain a single sentence. If a point requires further explanation, split it into separate items or move the detail into a sub-list.

#### Good Example (split into separate items)

```markdown
- The router forwards internal queries to the internal DNS server
- External queries go to upstream DNS servers
- DHCP distributes the router address as the DNS server for every VLAN
```

#### Good Example (sub-list for detail)

```markdown
- The RTX830 selectively forwards queries:
  - `home` domain and reverse lookups → internal DNS server
  - All other queries → upstream DNS servers
```

#### Bad Example

```markdown
- The router forwards internal queries to the internal DNS server. External queries go to upstream DNS servers obtained via DHCP on the WAN interface.
- DHCP distributes the router address as the DNS server for every VLAN. Each pool includes public DNS servers as fallbacks.
```

### 4. Empty Lines Around Lists

Always include exactly one empty line before and after lists. This applies to both ordered and unordered lists.

#### Good Example

```markdown
This is some text.

- Item 1
- Item 2
- Item 3

This is the next paragraph.
```

#### Bad Example

```markdown
This is some text.
- Item 1
- Item 2
- Item 3
This is the next paragraph.
```

Or

```markdown
This is some text.

- Item 1
- Item 2
- Item 3


This is the next paragraph.
```

## Rule Severity

- Empty Lines Around Headings: Warning
  - Affects readability, modification recommended
- Space Between Hash and Heading Text: Error
  - Critical for Markdown syntax, modification required
- One Sentence per List Item: Warning
  - Affects readability, modification recommended
- Empty Lines Around Lists: Warning
  - Affects readability, modification recommended
