# Mimsrv
An image server and UI for a simple web photo viewer

See my
[blog post](http://jim-mcbeath.blogspot.com/2018/03/golang-server-polymer-typescript-client.html)
for a discussion of why I chose to implement the server in Go and the
client in Polymer and Typescript.

## Quick Start

### Set up the environment

1. [Download](https://golang.org/dl/) and [install](https://golang.org/doc/install) Go
   * You may be able to install from your OS's system repository,
     if it is a new enough version; if not, use the
     download and install links on the above line
   * on Fedora run `sudo dnf install golang`
   * on debian run `sudo apt-get install golang`
1. Install git: `sudo dnf install git` or `sudo apt-get install git`
1. Install nodejs and npm using your package manager.
   * on debian, npm may not be in the apt repository, so:
       ```
       curl -sL https://deb.nodesource.com/setup_9.x | sudo -E bash -
       sudo apt-get install nodejs
       sudo apt-get install npm
       ```
1. Install typescript: `sudo npm install -g typescript`
1. Install bower: `sudo npm install -g bower`
1. Install polymer cli: `sudo npm install -g polymer-cli --unsafe-perm`

### Download mimsrv and dependencies

1. Set your GOPATH so `go get` knows where to put files:
   `export GOPATH=$HOME/go`
1. Download this repo and its go dependencies: `go get github.com/jimmc/mimsrv`
    * If you get an error, you may have an old version of `go`
1. cd into the mimsrv directory for the remaining work: `cd ~/go/src/github.com/jimmc/mimsrv`
1. Download polymer dependencies: `(cd _ui && bower install)`
1. If you want to view video files (mp4 and mpeg), install ffmpeg,
   such as (on Fedora): `sudo dnf install ffmpeg`

### Build and test

1. Build the server: `go build`
1. Run the server tests: `go test ./...`
1. Compile the UI typescript to javascript: `(cd _ui && tsc)`
1. Build the UI polymer bundle: `(cd _ui && polymer build)`

### Try it out

1. Run the demo server: `./demo`
1. Open [localhost:8021](http://localhost:8021/), login in as `user1` with password `pw1`,
   or as `editor` with password `pwe` to be able to use the editing features.
1. Click on the right-arrow to expand or collapse a folder;
   click on the three little horizontal lines to open the menu

## Data Format

Mimsrv is designed to serve image and text files from a directory hierarchy
rooted at a single location that is specified with the `--contentroot`
command line option.

Text files are only served if they have the extension `.txt`.
Image files are only served if they have one of the extensions
`.jpg`, `.jpeg`, `.png`, or `.gif`.

For each image file, the server looks for a text file that has the same
base name as the image file but with a `.txt` extension, and includes
the contents of that file as the descriptive text for the image.

For each directory, the server looks for a text file with the name
`summary.txt`, and includes the contents of that file as the
descriptive text for the directory.

The summary.txt file can include special directive lines that start
with an exclamation mark (!), followed by a command word:

*  ignoreFileTimes - do not display file times for this directory
*  sortByFileTimes - sort these files by modified time instead of name

Within each directory, the server looks for the file `index.mpr`
(mpr is for MimPRint) for meta-information about the images in the
directory. If that file exists, only images whose names are included
in that file (one image filename per line) will be displayed.
In addition, each image line in that file can include an optional
rotation flag that indicates if the image should be rotated by
90 (+r), 180 (+rr), or 270 (-r) degrees (or no rotation if none specified).
If there is no `index.mpr` file, then all images in the directory are
included when listing that directory, and no images are rotated.

The server assumes the timestamp on image files is the time that photo
was taken, and it returns that time with the meta-info for that image,
formatted in the local timezone, to be displayed in the list.
If the directory contains a file called TZ, the server assumes that is
a symlink to a timezone file, and it uses that timezone to format
the file timestamps for all of the image files in that directory.

## Authentication and Authorization

Mimsrv reads a simple password file in CSV format, specified with the
`--passwordfile` command line option, with one line per user.
The first field is the username, the second is the encrypted password,
and the third is the space-separated set of permissions.

An initial password file can be created by running mimsrv with the
`--createPasswordFile` option, then users can be added by running mimsrv
with the `--updatePasssword` option. In both cases you must also specify
the `--passwordfile` option. When either of these action options is used,
mimsrv exits after taking the requested action.

There is currently one permission defined: `edit`. This permission can
be manually added as the value of the third field for any user in the
password file.

All API calls (except for auth calls) require authentication, and
return an authorization error if the client is not authenticated. This
causes the UI code to show a login dialog. When the user logs in, the
server sends a cookie with an authentication token that is sent back on
subsequent calls, allowing the API to be accessed.
The cookie times out if no API calls are made in an hour, after which
the user must log in again.
It gets refreshed on every API call, but in any case expires ten hours after
initial authentication, after which the user must log in again.
The user can optionally log out at any time, in which case the
server clears the cookie from the client.

API requests that make changes, such as rotating a photo or updating
a description, require the `edit` permission.

Mimsrv uses a relatively simple standalone authentication system that
should be sufficient for casual protection. On login, the client code
collects a username and a password from the user. It combines the username
and password into a string that it then runs through sha256, resulting
in what is referred to in the code as the cryptword. It then gets the
current time in seconds since the epoch, formats it as a decimal string,
and combines that with the cryptword into a string that it then runs
through sha256 again. It passes the username, the epoch time, and the
encrypted time+cryptword to the server. The server looks up the username
in the password file, retrieves the cryptword from that record, checks
the time sent by the user to make sure it is within the specified clock skew
from the current time, combines the time and cryptword into a string that
it then runs through sha256, and compares that against the encrypted
time+cryptword that was sent by the client. If they match, the client
is authenticated.

With the above system, the user's password is never stored in cleartext
(the password file only stores the cryptword), and neither the password
nor the cryptword are ever sent over the wire. The current time is used in
the manner of a challenge token, but without requiring an additional round
trip, and prevents a replay attack outside of the clock skew window.

## Using the mimsrv UI

Start the server with the desired arguments to specify the password file,
the content root directory, and the port number (see the `demo` script
for an example), then point your browser to
that port number. You will need to log in using one of the accounts in
the password file.

At the main screen, a menu is available by clicking on the menu icon (three
little horizontal lines in a stack) at the top left corner.

You can expand or contract a directory in the nav bar on the left side of
the window by clicking on the right-arrow located to the left of the
directory name. The up and down arrows move the selection up and down in
the list. If a file is selected (as opposed to a directory), when you
move up or down and are at the end of the files in that directory, it
will move to the previous or next file, opening directories as necessary,
rather than moving to a directory.

Keyboard shortcuts can be used either when the nav bar has focus or when
the image area has focus, but you may need to open an image first in order
to get focus into the image area.

Use `?` or select Help from the menu to see the list of keyboard shortcuts.

## Video

Image listings in mimsrv can include mp4 and mpg files. When the client
requests an image for a video file, mimsrv runs ffmpeg to extract the first
frame of the video file as an image. The UI then overlays
a white "play" icon on top of that image. When the user clicks the play
icon, the UI client requests the video file from mimsrv.

When the client requests an mp4 file, mimsrv directly serves it, on the
assumption that it is a mp4/H.264-encoded video that modern browsers
understand. When the client requests an mpeg file, mimsrv assumes the
input file is mpeg-1 encoded, so calls out to ffmpeg to transcode it to
mp4/H.264. It saves the transcoded file in a cache directory `.mimcache`
within the directory containing the original mpeg file, so that the next
request for that mpeg video file can be served quickly from the previously
transcoded and cached file.

## About the Repository

When I first started on this project, I wasn't sure how to handle
having separate languages and two build processes for the server
and the client code, so I set up two repositories, one for mimsrv
and the other for the client code, mimview. Later, when I realized
I could just drop the mimview code into the mimsrv directory in a
subdirectory name starting with underscore so that the Go tools would ignore it,
I zipped together the two repositories as a single linear history in
commit-date order using some git magic so that
they are now a single repository.

This explains some of the oddities you might find in the early
commits, such as the fact that the first two commits both say they
are the initial commit:
[one](https://github.com/jimmc/mimsrv/commit/f7c7cf29d9e47b98aa26fbc2b23aa6ad4fa5a38e)
was the initial commit for mimsrv, the
[other](https://github.com/jimmc/mimsrv/commit/6a9c1172a70e2c6d23a362b0655c39f428c13105)
for mimview. There are also a few vestiges of mimview still
visible, such as the name of the command line option specifying the
location of the UI files.

There is a
[commit](https://github.com/jimmc/mimsrv/commit/525b53edc37dc5b9fc4645ef6e79b6c57128ec3c)
with comment <i>Minor cleanup after merging in mimview repo</i>
just after the point in time that the two histories
were merged (which unfortunately added an error such that it may not compile),
so you can use that commit to distinguish the earlier part of the project,
which was developed in two separate repositories, from the later part,
when both projects have been merged into one project in one repository.
