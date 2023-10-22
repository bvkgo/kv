// Copyright (c) 2023 BVK Chaitanya

package kvtests

const SerializedTxCommits = `
  db:db1  new-transaction                => tx:tx1
  tx:tx1  set key:0 value:zero
  tx:tx1  commit

  db:db1  new-transaction                => tx:tx2
  tx:tx2  get key:0                      => value:zero
  tx:tx2  set key:0 value:ZERO
  tx:tx2  commit

  db:db1  new-transaction                => tx:tx3
  tx:tx3  get key:0                      => value:ZERO
  tx:tx3  delete key:0
  tx:tx3  commit

  db:db1  new-transaction                => tx:tx4
  tx:tx4  get key:0                      => error:ErrNotExist
  tx:tx4  commit
`

const SerializedTxCommitsAndRollbacks = `
  db:db1  new-transaction                => tx:tx1
  tx:tx1  set key:0 value:zero
  tx:tx1  commit

  db:db1  new-transaction                => tx:tx2
  tx:tx2  get key:0                      => value:zero
  tx:tx2  set key:0 value:ZERO
  tx:tx2  rollback

  db:db1  new-transaction                => tx:tx3
  tx:tx3  get key:0                      => value:zero
  tx:tx3  delete key:0
  tx:tx3  rollback

  db:db1  new-transaction                => tx:tx4
  tx:tx4  get key:0                      => value:zero
  tx:tx4  delete key:0
  tx:tx4  commit

  db:db1  new-transaction                => tx:tx5
  tx:tx5  get key:0                      => error:ErrNotExist
  tx:tx5  set key:0 value:ZERO
  tx:tx5  commit

  db:db1  new-transaction                => tx:tx6
  tx:tx6  get key:0                      => value:ZERO
  tx:tx6  commit
`

const NonConflictingTxes = `
  db:db1  new-transaction               => tx:tx1
  db:db1  new-transaction               => tx:tx2
  db:db1  new-transaction               => tx:tx3

  tx:tx1  set key:1 value:one
  tx:tx2  set key:2 value:two
  tx:tx3  set key:3 value:three

  tx:tx1  commit
  tx:tx2  commit
  tx:tx3  commit
`

const ConflictingReadOnlyTxes = `
  db:db1  new-transaction               => tx:init
  tx:init set key:0 value:zero
  tx:init commit

  db:db1  new-transaction               => tx:tx1
  db:db1  new-transaction               => tx:tx2
  db:db1  new-transaction               => tx:tx3

  tx:tx1  get key:0                     => value:zero
  tx:tx2  get key:0                     => value:zero
  tx:tx3  get key:0                     => value:zero

  tx:tx1  commit
  tx:tx2  commit
  tx:tx3  commit
`

const ConflictingReadWriteTxes = `
  db:db1  new-transaction               => tx:init
  tx:init set key:0 value:zero
  tx:init commit

  db:db1  new-transaction               => tx:tx1
  db:db1  new-transaction               => tx:tx2

  tx:tx1  set key:0 value:ZERO
  tx:tx2  get key:0                     => value:zero

  tx:tx1  commit
  tx:tx2  commit                        => error:non-nil
`

const ConflictingDeletes = `
  db:db1  new-transaction               => tx:init
  tx:init set key:0 value:zero
  tx:init commit

  db:db1  new-transaction               => tx:tx1
  db:db1  new-transaction               => tx:tx2

  tx:tx1  delete key:0
  tx:tx1  set key:1 value:one

  tx:tx2  delete key:0
  tx:tx2  set key:2 value:two

  tx:tx1  commit
  tx:tx2  commit                        => error:non-nil
`

const NonConflictingDeletes = `
  db:db1  new-transaction               => tx:init
  tx:init set key:1 value:one
  tx:init set key:2 value:two
  tx:init commit

  db:db1  new-transaction               => tx:tx1
  db:db1  new-transaction               => tx:tx2

  tx:tx1  delete key:1
  tx:tx2  delete key:2

  tx:tx1  commit
  tx:tx2  commit
`

const AbortedReads = `
  db:db1  new-transaction               => tx:init
  tx:init set key:key value:value
  tx:init commit

  db:db1  new-transaction               => tx:tx1
  tx:tx1  set key:key value:VALUE
  tx:tx1  rollback

  db:db1  new-transaction               => tx:tx2
  tx:tx2  get key:key                   => value:value
  tx:tx2  commit
`

const RepeatedReads = `
  db:db1  new-transaction               => tx:init
  tx:init set key:key value:value
  tx:init commit

  db:db1  new-transaction               => tx:tx1
  db:db1  new-transaction               => tx:tx2

  tx:tx1  set key:key value:VALUE
  tx:tx1  commit

  tx:tx2  get key:key                   => value:value
  tx:tx2  commit                        => error:non-nil

  db:db1  new-transaction               => tx:tx3
  tx:tx3  get key:key                   => value:VALUE
  tx:tx3  commit
`
