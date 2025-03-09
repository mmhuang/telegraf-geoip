# Change log

本项目从[telegraf-geoip](https://github.com/a-bali/telegraf-geoip) folk出来。

背景：做一个基于用户的netflow的采集的分析系统。我的情况是内网IP对公网的访问。

netflow probe: linux server的flow采集使用[softflowd](https://github.com/irino/softflowd) netflow v9，后面也可以使用ipfix silk套餐的yaf，加上dpi和rf-ring，但编译的有点坑。
netflow colletor: 与telegraf的input.netflow input模块结合
    process.regex 用于上下行flow分流。
    process.lookup 用于通过my_ip_user.json, 将ip添加用户信息或设备信息。
                  【可以通过service telegraf reload 更新映射表，也就是对telegraf发送系统信号，或者改写一下lookup插件实现用户数据从数据库自动获取】
    process.geoip 也就是当前的项目，用于补充netflow的 geo信息和运营商信息。原来好像只可以二选一。这里的修改是对原来的geoip插件做修改，同时采集地理信息和运营商信息。
netflow store: influxdb v1.8（当前有些问题，做dst top20的分析会爆表，后面计划增加一些流计算，将常用的情况预先计算好），据说使用influxdb v2做聚合分析会更好一点，v3才可以解决tag 爆表的问题，也就是千亿级别的时间线问题，但是可惜还没有开源）
netflow show: 显示使用grafana，自己做一些图表，基操不表。

地理信息的库使用：
[github GeoLite2-Country](https://github.com/wp-statistics/GeoLite2-Country)
[github GeoLite2-City](https://github.com/wp-statistics/GeoLite2-Cit)
[github GeoLite2-ASN](https://github.com/wp-statistics/GeoLite2-ASN)

也可以从 [MaxMind GeoLite2](https://dev.maxmind.com/geoip/geoip2/geolite2/) 

以下是原来的使用帮助：

# Telegraf GeoIP processor plugin

This processor plugin for [telegraf](https://github.com/influxdata/telegraf) looks up IP addresses in the [MaxMind GeoLite2](https://dev.maxmind.com/geoip/geoip2/geolite2/) database and adds the respective ISO country code, city name, latitude and longitude as new fields to the output.

# Installation

This module is to be used as an external plugin to telegraf, therefore first compile it using Go:

    $ git clone https://github.com/a-bali/telegraf-geoip
    $ cd telegraf-geoip
    $ go build -o geoip cmd/main.go

This will create a standalone binary named `geoip`.

# Usage

You will need to add this plugin as an external plugin to your telegraf config as follows:

    [[processors.execd]]
    command = ["/path/to/geoip_binary", "--config", "/path/to/geoip_config_file"]

# Configuration

As specified above, the plugin uses a separate configuration file, where you can specify where it can find the downloaded GeoLite2 database (you will need the City version), which field to read as input and how to name the newly created fields. For details, please see the [sample config](https://github.com/a-bali/telegraf-geoip/blob/master/plugin.conf).

# License

This software is licensed under the MIT license.
