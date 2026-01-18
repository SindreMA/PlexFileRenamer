This application is a simple CLI tool that renames/move files in a directory based on the what show they are matched to in Plex.

Its written in golang, meant to be compiled to a single binary, and is currently only tested on Windows.

It should take the database path as argument.
It should support dry-run mode to preview changes without applying them.

Should support copy and move mode.

Plex database uses SQLite, so the application will need to read from the database file directly.

Should Go over all entries in the database.

For each root folder it checks, it should ask user if they want to rename the files.
Then for every show, if tv shows, it should ask user if they want to rename the files.
Also provide a option to approve all shows.


It should support custom name formats, and custom folder names.

Should support foldering based on seasons.

It should also support replcaing the path with a different path.

So if the media is in a network share, On the server it might be at f:\Media\Shows while on the client it might be at h:\Media\Shows.
Then the application should be able to rename the files to the correct path.

That way we don't need to to run it from the server directly.