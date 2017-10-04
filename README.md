[![Build Status](https://travis-ci.org/AdvancedClimateSystems/io.svg?branch=master)](https://travis-ci.org/AdvancedClimateSystems/io)

# IO

Go packages for pheripheral I/O. It contains driver for the following IC's:

* SPI
    * [Microchip][spi/microchip]
        * MCP3004
        * MCP3008
        * MCP3204
        * MCP3208
* I<sup>2</sup>C
    * [Maximum Integrated][i2c/max]
        * MAX5813
        * MAX5814
        * MAX5815
    * [Microchip][i2c/microchip]
        * MCP4725

    * [Texas Instruments][i2c/ti]
        * DAC5578
        * DAC6578
        * DAC7578

## License

IO is licensed under [Mozilla Public License][mpl] © 2017 [Advanced Climate
System][acs].

[acs]: http://advancedclimate.nl
[mpl]: LICENSE
[i2c/max]: https://godoc.org/github.com/AdvancedClimateSystems/io/i2c/max
[i2c/microchip]: https://godoc.org/github.com/AdvancedClimateSystems/io/i2c/microchip
[i2c/ti]: https://godoc.org/github.com/AdvancedClimateSystems/io/i2c/ti
[spi/microchip]: https://godoc.org/github.com/AdvancedClimateSystems/io/spi/microchip
