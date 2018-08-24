# Generating Animated Radars

## imagemagick

    convert -loop 0 out/*.png animated-fast.gif

## create animated gif and preview in terminal

    gor main.go -d ../../testdata/harvey -s 512 -p vel && convert out/*.png out/animated.png && imgcat animated.gif