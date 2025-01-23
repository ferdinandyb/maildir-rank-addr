# Changelog

## v1.4.0

 - multiple root mail folders can now be specified
 - does not look for maildir structure anymore, so any one-mail-per-file setup can be used
 - mbox files are also parsed
 - error messages now also show which file is causing them
 - Reply-To and Sender headers are also included in the analysis
 - List-Id is now used for figuring out the name of mailing lists (before it was easy to
   end up with "Maintainer via Listname" as the name of the list), the exact way this is
   done can be controlled via a template
