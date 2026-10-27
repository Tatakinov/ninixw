#!/usr/bin/ruby

request = [
  "GET SHIORI/3.0",
  'Charset: UTF-8',
  "Event: OnBoot",
]
request << "\r\n"
request = request.join("\r\n")
request = [[request.bytesize].pack('L'), request.force_encoding(Encoding::BINARY)].join

print request
