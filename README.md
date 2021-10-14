# BMO_JobqueueMonitor
Simple golang monitor for the jobqueue service running on bugzilla

Assumptions:
  - Config file must be named `jqm.json`
  - You have a statsd listener on `localhost:8125`

To Do's:
  - Accept command line parameters for help/config file location
  - Add line to config file to cover metric destination
  - Allow the json response to more arbitrary in length (more than 2 metrics)
  - Make this service agnostic; not just bugzilla
