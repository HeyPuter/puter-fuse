[
    {
        name: 'FAO',
        package: 'fao',

        imports: {
            // base: ['fmt'],
            interface: ['io'],
        },

        methods: {
            Stat: [
                [ ['path', 'string'] ],
                ['NodeInfo', 'bool', 'error'],
            ],
            ReadDir: [
                [ ['path', 'string'] ],
                ['[]NodeInfo', 'error']
            ],
            Read: [
                [ ['path', 'string'], ['dest', '[]byte'], ['off', 'int64'] ],
                ['int', 'error']
            ],
            Write: [
                [ ['path', 'string'], ['src', '[]byte'], ['off', 'int64'] ],
                ['int', 'error']
            ],
            Create: [
                [ ['path', 'string'], ['name', 'string'] ],
                ['NodeInfo', 'error']
            ],
            Truncate: [
                [ ['path', 'string'], ['size', 'uint64'] ],
                ['error']
            ],
            MkDir: [
                [ ['path', 'string'], ['name', 'string'] ],
                ['NodeInfo', 'error']
            ],
            Symlink: [
                [ ['parent', 'string'], ['name', 'string'], ['target', 'string'] ],
                ['NodeInfo', 'error']
            ],
            Unlink: [
                [ ['path', 'string'] ],
                ['error']
            ],
            Move: [
                [ ['source', 'string'], ['parent', 'string'], ['name', 'string'] ],
                ['error']
            ],
            // Copy: [
            //     [ ['source', 'string'], ['parent', 'string'], ['name', 'string'] ],
            //     ['NodeInfo', 'error']
            // ],
            ReadAll: [
                [ ['path', 'string'] ],
                ['io.ReadCloser', 'error']
            ]
        }
    }
]
