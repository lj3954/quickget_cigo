[![Config generation status](https://github.com/lj3954/quickget_cigo/actions/workflows/run.yml/badge.svg)](https://lj3954.github.io/quickget_cigo/)

# Quickget CIgo

This project is responsible for generating configuration files containing download links, various information such as OS homepages, and
other data required for creating virtual machines from the images.

It's behind the scenes of [quickemu-rs](https://github.com/lj3954/quickemu-rs)'s quickget, along with [quickosdl](https://github.com/lj3954/quickosdl), among others.

Data is published daily in JSON format, and can easily be included within other projects. Currently, the formatting is unstable and subject to change.
