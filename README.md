# LEC3 Converter

`lec3-conv` converts the scanned images to fit the specific e-book device.
Each supported device configurations are located in `./config` directory.

Image conversion processes are:
1. Adjust the text line spacing to match the resolution ratio of the specified e-book device.
2. Resize image to match the resolution of the e-book device.

## Usage
```bash
$ lec3-conv [-cfg=$(configFilePath)] [-src=$(source directory)] [-dest=$(destination directory)]
```

For example, to convert images in `./input` directory to `./output` directory using `./config/koboAuraHD-cbz.yaml` configuration file:

```bash
$ lec3-conv -cfg=./config/koboAuraHD-cbz.yaml -src=./input -dest=./output
```
