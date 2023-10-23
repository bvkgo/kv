// Copyright (c) 2023 BVK Chaitanya

package kvtests

var BasicOpsTemplateMap = map[string]string{
	"GetSetDelete": `
  db:db1  new-transaction       => tx:tx1

  tx:tx1  delete key:0          => error:nil|ErrNotExist

  tx:tx1  get key:0             => error:ErrNotExist
  tx:tx1  set key:0 value:zero
  tx:tx1  get key:0             => value:zero

  tx:tx1  delete key:0
  tx:tx1  get key:0             => error:ErrNotExist

  tx:tx1  commit
`,

	"AscendDescendScanEmpty": `
  db:db1  new-transaction       => tx:tx1

  tx:tx1  ascend begin: end:    => it:it1
  it:it1  current               => key: value: error:nil
  it:it1  next                  => key: value: error:nil

  tx:tx1  descend begin: end:   => it:it2
  it:it2  current               => key: value: error:nil
  it:it2  next                  => key: value: error:nil

  tx:tx1  scan                  => it:it3
  it:it3  current               => key: value: error:nil
  it:it3  next                  => key: value: error:nil

  tx:tx1  commit
`,

	"AscendDescendInvalid": `
  db:db1  new-transaction       => tx:tx1

  tx:tx1  set key:0 value:zero
  tx:tx1  set key:1 value:one
  tx:tx1  set key:2 value:two

  tx:tx1  ascend begin:2 end:1    => error:ErrInvalid
  tx:tx1  descend begin:2 end:1   => error:ErrInvalid

  tx:tx1  ascend begin:2 end:0    => error:ErrInvalid
  tx:tx1  descend begin:2 end:0   => error:ErrInvalid

  tx:tx1  ascend begin:1 end:0    => error:ErrInvalid
  tx:tx1  descend begin:1 end:0   => error:ErrInvalid

  tx:tx1  rollback
`,

	"AscendEmpty": `
  db:db1  new-transaction       => tx:tx1

  tx:tx1  set key:0 value:zero
  tx:tx1  set key:1 value:one
  tx:tx1  set key:2 value:two

  tx:tx1  ascend begin:0 end:0  => it:it1
  it:it1  current               => key: value: error:nil
  it:it1  next                  => key: value: error:nil

  tx:tx1  ascend begin:1 end:1  => it:it2
  it:it2  current               => key: value: error:nil
  it:it2  next                  => key: value: error:nil

  tx:tx1  ascend begin:2 end:2  => it:it3
  it:it3  current               => key: value: error:nil
  it:it3  next                  => key: value: error:nil

  tx:tx1  rollback
`,

	"DescendEmpty": `
  db:db1  new-transaction         => tx:tx1

  tx:tx1  set key:0 value:zero
  tx:tx1  set key:1 value:one
  tx:tx1  set key:2 value:two

  tx:tx1  descend begin:0 end:0  => it:it1
  it:it1  current                => key: value: error:nil
  it:it1  next                   => key: value: error:nil

  tx:tx1  descend begin:1 end:1  => it:it2
  it:it2  current                => key: value: error:nil
  it:it2  next                   => key: value: error:nil

  tx:tx1  descend begin:2 end:2  => it:it3
  it:it3  current                => key: value: error:nil
  it:it3  next                   => key: value: error:nil

  tx:tx1  rollback
`,

	"AscendNonEmptyRange": `
  db:db1  new-transaction         => tx:tx1

  tx:tx1  set key:0 value:zero
  tx:tx1  set key:1 value:one
  tx:tx1  set key:2 value:two
  tx:tx1  set key:3 value:three
  tx:tx1  set key:4 value:four

  tx:tx1  ascend begin:0 end:5  => it:it1
  it:it1  current               => key:0 value:zero
  it:it1  next                  => key:1 value:one
  it:it1  next                  => key:2 value:two
  it:it1  next                  => key:3 value:three
  it:it1  next                  => key:4 value:four
  it:it1  next                  => key: value: error:nil

  tx:tx1  ascend begin:0 end:4  => it:it2
  it:it2  current               => key:0 value:zero
  it:it2  next                  => key:1 value:one
  it:it2  next                  => key:2 value:two
  it:it2  next                  => key:3 value:three
  it:it2  next                  => key: value: error:nil

  tx:tx1  ascend begin:0 end:1  => it:it3
  it:it3  current               => key:0 value:zero
  it:it3  next                  => key: value: error:nil

  tx:tx1  ascend begin:1 end:2  => it:it4
  it:it4  current               => key:1 value:one
  it:it4  next                  => key: value: error:nil

  tx:tx1  ascend begin:2 end:25  => it:it5
  it:it5  current                => key:2 value:two
  it:it5  next                   => key: value: error:nil

  tx:tx1  ascend begin:25 end:35  => it:it6
  it:it6  current                 => key:3 value:three
  it:it6  next                    => key: value: error:nil

  tx:tx1  rollback
`,

	"DescendNonEmptyRange": `
  db:db1  new-transaction         => tx:tx1

  tx:tx1  set key:0 value:zero
  tx:tx1  set key:1 value:one
  tx:tx1  set key:2 value:two
  tx:tx1  set key:3 value:three
  tx:tx1  set key:4 value:four

  tx:tx1  descend begin:0 end:5  => it:it1
  it:it1  current                => key:4 value:four
  it:it1  next                   => key:3 value:three
  it:it1  next                   => key:2 value:two
  it:it1  next                   => key:1 value:one
  it:it1  next                   => key:0 value:zero
  it:it1  next                   => key: value: error:nil

  tx:tx1  descend begin:0 end:4  => it:it2
  it:it2  current                => key:3 value:three
  it:it2  next                   => key:2 value:two
  it:it2  next                   => key:1 value:one
  it:it2  next                   => key:0 value:zero
  it:it2  next                   => key: value: error:nil

  tx:tx1  descend begin:0 end:1  => it:it3
  it:it3  current                => key:0 value:zero
  it:it3  next                   => key: value: error:nil

  tx:tx1  descend begin:1 end:2  => it:it4
  it:it4  current                => key:1 value:one
  it:it4  next                   => key: value: error:nil

  tx:tx1  descend begin:2 end:25  => it:it5
  it:it5  current                 => key:2 value:two
  it:it5  next                    => key: value: error:nil

  tx:tx1  descend begin:25 end:35  => it:it6
  it:it6  current                  => key:3 value:three
  it:it6  next                     => key: value: error:nil

  tx:tx1  rollback
`,

	"AscendOneEmptyRange": `
  db:db1  new-transaction         => tx:tx1

  tx:tx1  set key:0 value:zero
  tx:tx1  set key:1 value:one
  tx:tx1  set key:2 value:two
  tx:tx1  set key:3 value:three
  tx:tx1  set key:4 value:four

  tx:tx1  ascend begin: end:5     => it:it1
  it:it1  current                 => key:0 value:zero
  it:it1  next                    => key:1 value:one
  it:it1  next                    => key:2 value:two
  it:it1  next                    => key:3 value:three
  it:it1  next                    => key:4 value:four
  it:it1  next                    => key: value: error:nil

  tx:tx1  ascend begin: end:4     => it:it2
  it:it2  current                 => key:0 value:zero
  it:it2  next                    => key:1 value:one
  it:it2  next                    => key:2 value:two
  it:it2  next                    => key:3 value:three
  it:it2  next                    => key: value: error:nil

  tx:tx1  ascend begin: end:2     => it:it3
  it:it3  current                 => key:0 value:zero
  it:it3  next                    => key:1 value:one
  it:it3  next                    => key: value: error:nil

  tx:tx1  ascend begin:0 end:     => it:it4
  it:it4  current                 => key:0 value:zero
  it:it4  next                    => key:1 value:one
  it:it4  next                    => key:2 value:two
  it:it4  next                    => key:3 value:three
  it:it4  next                    => key:4 value:four
  it:it4  next                    => key: value: error:nil

  tx:tx1  ascend begin:2 end:     => it:it4
  it:it4  current                 => key:2 value:two
  it:it4  next                    => key:3 value:three
  it:it4  next                    => key:4 value:four
  it:it4  next                    => key: value: error:nil

  tx:tx1  rollback
`,

	"DescendOneEmptyRange": `
  db:db1  new-transaction         => tx:tx1

  tx:tx1  set key:0 value:zero
  tx:tx1  set key:1 value:one
  tx:tx1  set key:2 value:two
  tx:tx1  set key:3 value:three
  tx:tx1  set key:4 value:four

  tx:tx1  descend begin: end:5     => it:it1
  it:it1  current                 => key:4 value:four
  it:it1  next                    => key:3 value:three
  it:it1  next                    => key:2 value:two
  it:it1  next                    => key:1 value:one
  it:it1  next                    => key:0 value:zero
  it:it1  next                    => key: value: error:nil

  tx:tx1  descend begin: end:4     => it:it2
  it:it2  current                 => key:3 value:three
  it:it2  next                    => key:2 value:two
  it:it2  next                    => key:1 value:one
  it:it2  next                    => key:0 value:zero
  it:it2  next                    => key: value: error:nil

  tx:tx1  descend begin: end:2     => it:it3
  it:it3  current                 => key:1 value:one
  it:it3  next                    => key:0 value:zero
  it:it3  next                    => key: value: error:nil

  tx:tx1  descend begin:0 end:     => it:it4
  it:it4  current                 => key:4 value:four
  it:it4  next                    => key:3 value:three
  it:it4  next                    => key:2 value:two
  it:it4  next                    => key:1 value:one
  it:it4  next                    => key:0 value:zero
  it:it4  next                    => key: value: error:nil

  tx:tx1  descend begin:2 end:     => it:it4
  it:it4  current                 => key:4 value:four
  it:it4  next                    => key:3 value:three
  it:it4  next                    => key:2 value:two
  it:it4  next                    => key: value: error:nil

  tx:tx1  rollback
`,
}
