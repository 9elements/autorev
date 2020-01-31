# Database

The database consists of three tables: _tests_, _traceLog_ and _updOptions_. The
table _updOptions_ consists all upd options which can be set by the Host
program. The table tests inherits the tests which should be run on the host. It
contains one column _status_ which can be:

* 0 = Test has not been run yet
* 1 = Test is in progress right now
* 2 = Test has been run successfully
* 3 = Test failed for some reason.

Also the table tests contains the timestamp when the test has been created, has
been started and has been finished. It also contains the complete Log output in
the _completeLog_ column. 
