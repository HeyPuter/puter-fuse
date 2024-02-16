[
    {
        name: 'FAO',
        package: 'fao',

        methods: {
            Stat: [
                [ ['path', 'string'] ],
                ['NodeInfo', 'error'],
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
            Truncate: [
                [ ['path', 'string'], ['size', 'int64'] ],
                ['error']
            ],
            Link: [
                [ ['parent', 'string'], ['name', 'string'], ['target', 'string'] ],
                ['error']
            ],
            ReadAll: [
                [ ['path', 'string'] ],
                ['[]byte', 'error'],
                `
                stat, err := base.Stat(path)
                if err != nil {
                    return nil, err
                }
                buf := make([]byte, stat.Size)
                n, err := base.Read(path, buf, 0)
                if err != nil {
                    return nil, err
                }
                return buf[:n], nil
                `
            ]
        }
    }
]
