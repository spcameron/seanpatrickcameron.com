# Markdown Rules: CommonMark-ish

## Headers

A line is recognized as a header if and only if the following is true:
- **Indentation**: The line may begin with 0-3 spaces. Tabs do not count as indentation.
- **Marker run**: After leading spaces, there is a a run of 1-6 `#` characters.
- **Delimiter**: The marker run is followed by at least one delimiter character: space or tab.
- **Content**: Header content is the rest of the line after consuming all consecutive spaces or tabs following the marker run.
- **Normalization**: Trailing whitespace is trimmed from the content.
- **Termination**: The header is a single line. A newline ends it.

