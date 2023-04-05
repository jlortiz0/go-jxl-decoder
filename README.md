# go-jxl-decode

A Golang wrapper for (the decoding part of) libjxl. Made in 3 hours as an experiment.

## Building

On Windows, download the latest release of [libjxl](https://github.com/libjxl/libjxl) and extract the DLLs to the same directory as your application. You need all of them.

After building, the application might be statically linked? I'm not sure about that but it seems to be the case.

On Linux, install `libbrotli-dev` and, if your distro has it, `libjxl-dev`. (If not, get the package from the above link.)

## Usage

This library registers itself with `image` and additionally exports `Decode` and `DecodeConfig`, which work as you might expect.

Note that only `Gray`, `RGBA`, and `NRGBA` color models and their 16-bit counterparts are identitifed by the library.
