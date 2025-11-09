# Sculptor 

Sculptor is a program for generating utility CSS files. Look at the `theme.json` file to configure.

## Available options:

- text colors
- background colors
- border colors
- margins
- padding
- border width

## Usage

```
sculptor -config [JSON config] -output [CSS output]
```

ex.
```
sculptor -config theme.json -output style.css
```

Use this directly, or `@import` into another CSS file.
