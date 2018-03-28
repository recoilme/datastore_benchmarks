# Datastore benchmarks

A very simple and rudimentary set of test to assess the performance gain of the transition from `flatfs` to `badger`.

What kind of use cases (read/write profiles) are of interest?

* Sequential reads of a single file.
* Random reads.
* Writes with GC.

Interact directly with the data store through [`go-datastore`](/cmd/datastore/main.go) and high level usage through the [Core API](/cmd/coreapi/main.go).


Value disk I/O Vs SSTables I/O
==============================

How big are the tables? I need to make sure they are being accessed from disk not memory (big enough DB), if not, only value retrieval is going to dominate.

Profile
=======

More reads than writes, this is what LSM has troubles with


Parallel SSD fetches?
===============


Fragmentation
=============
