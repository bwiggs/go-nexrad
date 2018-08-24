# Generating Radar Products

Products are what we know as radar images. Currently there are two supported products. Reflectivity and Velocity.

## Nexrad Level II Data Files

You will need the raw nexrad data files to process into radar products. Since they're stored on AWS S3, it's easiest to use the aws-cli tools to download them.

If you want to do an animated gif you'll need multiple data files

### Hurricane Harvey - Full Force

    $ aws s3 cp s3://noaa-nexrad-level2/2017/08/25/KCRP/KCRP20170825_235733_V06 .

### Hurricane Harvey - Landfall 2017/8/25

This will copy down all the data for August 25 from the KCRP radar site. WARNING: This is going to be big, 2GB+

    $ aws s3 cp --recursive s3://noaa-nexrad-level2/2017/08/25/KCRP KCRP

# Generate image(s) from the data file

The default output file is `radar.png`

    $ nexrad-render -f KCRP20170825_235733_V06

To process an entire directory. Default output to `./out/`

    $ nexrad-render -d KCRP

## create animated gif

Once you have a directory of products, use `imagemagick` to create an animated gif.

    $ convert -loop 0 out/*.png animated.gif

### terminal preview

Some terminals support viewing images in them. Use `imgcat` to view.

    $ imgcat animated.gif

## All at once

Generate an animated velocity radar image and preview

    $ nexrad-render -d KCRP -s 512 -p vel && convert out/*.png out/animated.png && imgcat animated.gif