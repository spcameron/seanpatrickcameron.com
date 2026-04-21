---
title: "Test Post"
slug: "test-post"
date: "1987-06-21"
---
# Header 1
## Header 2
### Header 3
#### Header 4
##### Header 5
###### Header 6

This header is underlined by setext (equals).
===

This header is underlined by setext (dashes).
---

This is a paragraph.

This line has *emphasis* and **strong** text, and even ***strong emphasis***.

This line has _underscore emphasis_ and __underscore strong__ text.

This line has mixed delimiters like **strong *nested emphasis*** inside.

This line has a [link to Google.](https://google.com "link to google")

This line has an image: ![alt text](image.png "image title")

This line has an autolink: <https://google.com>

This line has an email autolink: <test@example.com>

This line has a code span: `inline code`

This line has a longer code span: `` code with `backticks` inside ``

This is a paragraph
with a soft break (\n).

This is a paragraph

with a hard break (\n\n).

This line
- - -
is separated
***
by thematic breaks.
_ _ _

> This is a block quote.
> > This is a nested block quote.

> This is a block quote
>
> separated by a blank line.

- This is an unordered list
- This is the second list item
    - This is a nested list
    - With a second list item
- This is the third list item

1. This
2. Is
3. An
    - Interrupting
    - With an
    - Unordered list
4. Ordered
5. List

```go
fmt.Println("This is a backtick-fenced code block").
```

~~~go
fmt.Println("This is a tilde-fenced code block.")
~~~

```
x := 1
y := 2

if x == 1 {
    fmt.Println("This code block that preserves internal indentation.")
}
```

```html
<p>This code block contains HTML.</p>
<div>The HTML is still rendered literally.</div>
```

    This is an indented code block
    
    containing a blank line.

This is a normal paragraph line.

<!-- this is a comment that renders literally in the HTML -->

---

# ###

# Header with closer ###

# Header with mismatched closer ##

# Header### (no separator, hashes should remain)

# Header ### trailing text (closer should not apply)

---

\*this is not emphasis\*

\# this is not a header

This has an escaped link: \[not a link](https://example.com)

This has an escaped image: \![not an image](image.png)

This has escaped punctuation: \* \_ \# \[ \]

---

This has `code with *emphasis* inside` that should not parse emphasis.

This has **strong with `code span` inside**.

This has a [link with `code span` inside](https://example.com).

This has an ![image with *emphasis* inside](image.png).

---

[outer [inner] still outer](https://example.com)

[link with emphasis *inside* label](https://example.com)

[link with strong **inside** label](https://example.com)

[link with nested link [illegal](https://inner.com)](https://outer.com)

---

<https://example.com>

<https://example.com/test?query=1>

<test@example.com>

<invalid@>

<not a link>

---

This has inline HTML: <span>inline</span> content.

This has a tag with attributes: <a href="https://example.com">link</a>.

<div>
This is an HTML block that should terminate on a blank line.
</div>

This is after the HTML block.

---

This is a paragraph
- that should interrupt into a list

This is a paragraph

- that should clearly be a list

---

> - List inside block quote
> - Second item
>
> > Nested quote with list
> > - Item

---

> ### Header in quote ###
>
> Paragraph with *emphasis*, a [link](https://example.com), and `code`.
>
> - List item
>     - Nested with **strong**
>
> <span>inline HTML</span>
