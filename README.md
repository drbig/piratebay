# piratebay [![Build Status](https://travis-ci.org/drbig/piratebay.svg?branch=master)](https://travis-ci.org/drbig/piratebay) [![Coverage Status](https://img.shields.io/coveralls/drbig/piratebay.svg)](https://coveralls.io/r/drbig/piratebay?branch=master) [![GoDoc](https://godoc.org/github.com/drbig/piratebay?status.svg)](http://godoc.org/github.com/drbig/piratebay)

A Go library implementing robust and comprehensive searching of PirateBay.

Features:

- Regexp-based scraping with careful abstractions
- Automatic discovery of torrent categories and sort orders
- Leverage sorting on PirateBay's side
- Stratified fetching and parsing of details
- From basic search result down to file details per torrent
- Extensible filters framework
- Currently filters for: seeders, leechers, total size, file names
- Static and live test suite (needs more love though)
- Pure Go, no additional dependencies
- Includes a minimal command line interface example

## Showcase

Using the included *demo* `pbcmd`:

    $ ./pbcmd
    Usage: ./pbcmd [options] query query...
    
    Won't run any queries if any of -sf, -so, and -sc options have been supplied.
    Running a query or using -so or -sc requires a connection to PirateBay.
    
      -c="all": category filter ('unique category' or 'group/category')
      -d=false: print details for each torrent
      -debug=false: enable library debug output
      -f=false: only print first match
      -filters="": filters to apply (in sequence)
      -m=false: only print magnet link
      -o="seeders": sorting order (always descending)
      -sc=false: fetch and print available categories
      -sf=false: print available filters
      -so=false: fetch and print available orderings
      -version=false: show version and exit

- - -

    $ ./pbcmd -sf
    Available filters:
    files(include - regexp | exclude - regexp) - Filter by torrent files' name include/exclude
    seeders(min - int | max - int) - Filter by torrent min/max seeders
    leechers(min - int | max - int) - Filter by torrent min/max leechers
    size(min - int | max - int) - Filter by torrent total min/max size

- - -

    $ ./pbcmd -so
    Available sort orders:
    type
    seeders
    name
    uploaded
    size
    uled by
    leechers

- - -

    $ ./pbcmd -sc
    Available categories:
    games/pc
    games/mac
    games/psx
    games/handheld
    games/ios (ipad/iphone)
    games/other
    games/xbox
    games/wii
    games/android
    porn/movies
    porn/movies dvdr
    porn/pictures
    porn/games
    porn/hd - movies
    porn/movie clips
    porn/other
    other/e-books
    other/comics
    other/pictures
    other/covers
    other/physibles
    other/other
    /all
    audio/music
    audio/audio books
    audio/sound clips
    audio/flac
    audio/other
    video/
    video/movies dvdr
    video/handheld
    video/hd - movies
    video/tv shows
    video/hd - tv shows
    video/other
    video/movies
    video/music videos
    video/movie clips
    applications/other os
    applications/windows
    applications/mac
    applications/unix
    applications/handheld
    applications/ios (ipad/iphone)
    applications/android

- - -

    $ ./pbcmd -d -filters "seeders:min:1;files:exclude:.*\\.rar;files:include:.*\\.iso" -c unix freebsd
     1  1  FreeBSD-9.1-RELEASE-i386-dvd1.iso                                   
           2.31 GiB    2013-01-14 19:04:49 GMT  L33ch88                        
           http://thepiratebay.org/torrent/8020414
           Files: _____________________________________________________________
        1  FreeBSD-9.1-RELEASE-i386-dvd1.iso                             2.35 G
    
     1  2  FreeBSD-10.0-i386                                                   
           3.49 GiB    2014-05-16 08:42:36 GMT  Anonymous                      
           http://thepiratebay.org/torrent/10159470
           Files: _____________________________________________________________
        1  CHECKSUM.MD5-10.0-RELEASE-i386                                   313
        2  FreeBSD-10.0-RELEASE-i386-bootonly.iso                      194.81 M
        3  FreeBSD-10.0-RELEASE-i386-disc1.iso                         563.92 M
        4  FreeBSD-10.0-RELEASE-i386-dvd1.iso                            2.16 G
        5  FreeBSD-10.0-RELEASE-i386-memstick.img                      601.27 M
    
     1  3  FreeBSD 9.0 DVD 64-bit                                              
           2.22 GiB    2012-03-09 18:21:12 GMT  nldsfim                        
           http://thepiratebay.org/torrent/7089521
           Files: _____________________________________________________________
        1  FreeBSD-9.0-RELEASE-amd64-dvd1.iso                            2.22 G
    
     1  4  FreeBSD 9.1-RELEASE AMD64 DVD1                                      
           2.49 GiB    2013-01-30 21:17:54 GMT  manalishi666                   
           http://thepiratebay.org/torrent/8083961
           Files: _____________________________________________________________
        1  FreeBSD-9.1-RELEASE-amd64-dvd1.iso                            2.49 G
    
     1  5  FreeBSD 7.2 PowerPc All Media                                       
           768.38 MiB  2009-10-29 20:31:26 GMT  lan3y                          
           http://thepiratebay.org/torrent/5140205
           Files: _____________________________________________________________
        1  /7.2-RELEASE-powerpc-bootonly.iso                            27.85 M
        2  /7.2-RELEASE-powerpc-disc1.iso                              418.64 M
        3  /7.2-RELEASE-powerpc-docs.iso                                321.9 M
    
     1  6  FreeBSD-10.0-ia64                                                   
           1.8 GiB     2014-05-16 08:50:50 GMT  Anonymous                      
           http://thepiratebay.org/torrent/10159482
           Files: _____________________________________________________________
        1  CHECKSUM.MD5-10.0-RELEASE-ia64                                   237
        2  FreeBSD-10.0-RELEASE-ia64-bootonly.iso                      372.07 M
        3  FreeBSD-10.0-RELEASE-ia64-disc1.iso                         777.88 M
        4  FreeBSD-10.0-RELEASE-ia64-memstick.img                      694.27 M
    
     1  7  FreeBSD-10.0-amd64                                                  
           3.78 GiB    2014-05-16 08:34:08 GMT  Anonymous                      
           http://thepiratebay.org/torrent/10159453
           Files: _____________________________________________________________
        1  CHECKSUM.MD5-10.0-RELEASE-amd64                                  317
        2  FreeBSD-10.0-RELEASE-amd64-bootonly.iso                     209.96 M
        3  FreeBSD-10.0-RELEASE-amd64-disc1.iso                        622.75 M
        4  FreeBSD-10.0-RELEASE-amd64-dvd1.iso                           2.31 G
        5  FreeBSD-10.0-RELEASE-amd64-memstick.img                     665.08 M
    
     1  8  FreeBSD 9.2-RELEASE AMD64 DVD1                                      
           2.38 GiB    2013-10-12 19:15:54 GMT  Anonymous                      
           http://thepiratebay.org/torrent/9040079
           Files: _____________________________________________________________
        1  FreeBSD-9.2-RELEASE-amd64-dvd1.iso                            2.38 G
    
     1  9  FreeBSD 8.2 DVD 64-bit                                              
           2.25 GiB    2012-03-09 18:16:42 GMT  nldsfim                        
           http://thepiratebay.org/torrent/7089511
           Files: _____________________________________________________________
        1  FreeBSD-8.2-RELEASE-amd64-dvd1.iso                            2.25 G
    
     1 10  FreeBSD 9.0 ALL 64-bit                                              
           3.6 GiB     2012-03-07 18:46:45 GMT  nldsfim                        
           http://thepiratebay.org/torrent/7084752
           Files: _____________________________________________________________
        1  FreeBSD-9.0-RELEASE-amd64-bootonly.iso                      138.88 M
        2  FreeBSD-9.0-RELEASE-amd64-disc1.iso                         612.06 M
        3  FreeBSD-9.0-RELEASE-amd64-dvd1.iso                            2.22 G
        4  FreeBSD-9.0-RELEASE-amd64-memstick.img                      653.94 M
    
     1 11  FreeBSD 9.0 CD 64-bit                                               
           612.06 MiB  2012-03-09 18:20:28 GMT  nldsfim                        
           http://thepiratebay.org/torrent/7089520
           Files: _____________________________________________________________
        1  FreeBSD-9.0-RELEASE-amd64-disc1.iso                         612.06 M

- - -

    $ ./pbcmd -f -m -filters seeders:min:5 -c unix linux
    magnet:?xt=urn:btih:eb851063bbbe8ec29fe2ebbdc19d90df982eaeff&dn=Linux+-+Google+Chromium+i686-2.4.1290+ISO&tr=udp%3A%2F%2Ftracker.openbittorrent.com%3A80&tr=udp%3A%2F%2Ftracker.publicbt.com%3A80&tr=udp%3A%2F%2Ftracker.istole.it%3A6969&tr=udp%3A%2F%2Fopen.demonii.com%3A1337

Bonus `getlastep` command:

    $ ./getlastep
    Usage: ./getlastep [options...] show show...
    
      -c="": full Transmission RPC URL
      -v=false: print version and exit

- - -

    $ ./getlastep -c http://user:pass@host:port House
     1. House [2004 - Canceled/Ended]
        S08E22 "Everybody Dies" (2012-21-05, 910 days ago)
      - Found you a torrent:
        House S08E22 720p HDTV x264-DIMENSION [eztv] (7288837)
        Added torrent :)

## Development

### General notes

Running `go test` will *make requests to PirateBay*. Running `go test -short` will only run the fake-data tests, hence no network needed. From library standpoint the test suite needs most love.

Current test coverage is between around 51% - 59%, for fake-data-only and full tests respectively.

~~I have a todo item to get myself acquainted with at least one CI system available out there, and this project tops the list as it's a library.~~

Right after tests is the topic of the filters framework. I'd like them to be easy to add, maybe even loadable at runtime (no idea how to approach that yet). Feel free to go wild here.

### Some rationale

Given that web scraping is always akin to shooting at a moving target I've decided that regexp-based approach is the best for long-term (and I do have some experience here). Current approach makes small adjustments easy - just hammer `fragile.go` until it works. Only if PirateBay changes the layout radically one should be forced to revise `parsing.go`.

As much as I appreciate XPath and more semantic approach I don't think it is appropriate in this case.

## Contributing

Follow the usual GitHub development model:

1. Clone the repository
2. Make your changes on a separate branch
3. Make sure you run `gofmt` and `go test` before committing
4. Make a pull request

See licensing for legalese.

## Moral notes

Leaving the copyright law outside: support financially the stuff you care about with the same rationale that you support free software, sharing and the general idea of not being an asshole. All digital creations are still ultimately the result of work of other people, remember to keep that in mind.

## Licensing

Standard two-clause BSD license, see LICENSE.txt for details.

Any contributions will be licensed under the same conditions.

Neither I nor any contributor take any responsibility for the usage of this software, the same way that kitchen utensils manufacturers don't take responsibility for stabbings.

Copyright (c) 2014 Piotr S. Staszewski
