# Assets

## Favicon

Created with Inkscape.

The key inside the coin is based on [this](https://openclipart.org/detail/3330/key) clipart made by [barretr](https://openclipart.org/artist/barretr)
and hosted on [Open Clipart](https://openclipart.org/) licensed under the [Creative Commons Zero 1.0 License](https://creativecommons.org/publicdomain/zero/1.0/).

To work on the favicon, import it into Inkscape. Make sure to always save
to an "Inkscape SVG" to persist Inkscape specific stuff like layers.

To export the favicon: Select all objects and make sure to only export
the selected objects excluding the page.

For creating the actual favicon I use [RealFaviconGenerator](https://realfavicongenerator.net/).
Here I use my exported SVG as an input to create favicons in different formats.

## Swagger Initializer

The file [`./swagger-initializer.js`](./swagger-initializer.js) is used to
override the file with the same name that comes with the default Swagger UI
distribution. It is injected by the [`../scripts/place-swagger-ui`](../scripts/place-swagger-ui)
script.
