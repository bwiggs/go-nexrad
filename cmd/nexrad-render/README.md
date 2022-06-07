# Usage

    $ ./nexrad-render -h
    nexrad-render generates products from NEXRAD Level 2 (archive 2) data files.

    Usage:
    nexrad-render [flags]

    Flags:
    -c, --color-scheme string   color scheme to use. noaa, scope, pink (default "noaa")
    -d, --directory string      directory of L2 files to process
    -f, --file string           archive 2 file to process
    -h, --help                  help for nexrad-render
    -l, --log-level string      log level, debug, info, warn, error (default "warn")
    -o, --output string         output radar image
    -p, --product string        product to produce. ex: ref, vel (default "ref")
    -s, --size int32            size in pixel of the output image (default 1024)

# Installation

```
go install github.com/bwiggs/go-nexrad/cmd/nexrad-render@latest
```

# Generating Radar Products

Products are what we know as radar images. Currently there are two supported products. Reflectivity and Velocity.

## Nexrad Level II Data Files

You will need the raw nexrad data files to process into radar products. Since they're stored on AWS S3, it's easiest to use the aws-cli tools to download them.

If you want to do an animated gif you'll need multiple data files

### Single File - Hurricane Harvey - Full Force

    $ aws s3 cp s3://noaa-nexrad-level2/2017/08/25/KCRP/KCRP20170825_235733_V06 .

### Multiple Files - Hurricane Harvey - Landfall 2017/8/25

This will copy down all the data for August 25 from the KCRP radar site. WARNING: This is going to be big, 2GB+

    $ aws s3 cp --recursive s3://noaa-nexrad-level2/2017/08/25/KCRP KCRP

## Generate products from the data file(s)

The default output file is `radar.png` This will create a single product output.

    $ nexrad-render -f KCRP20170825_235733_V06

To process an entire directory. Default output to `./out/`

    $ nexrad-render -d KCRP

## Animated Gifs

Once you have a directory of products, use `imagemagick` to create an animated gif.

    $ convert -loop 0 out/*.png animated.gif

### terminal preview

Some terminals support viewing images in them. Use `imgcat` to view.

    $ imgcat animated.gif

## One Liner

Generate an animated velocity radar image and preview in terminal

    $ nexrad-render -d KCRP -s 512 -p vel && convert out/*.png out/animated.png && imgcat animated.gif