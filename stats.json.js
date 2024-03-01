({
    'read tests': [
        {
            a: 'initial test',
            test: 'node example.js',
            _: [
                ['2m20.460s', { first_run: true }], // TODO: re-test
                ['2m18s'],
            ]
        },
        {
            a: 'after adding no_thumbs and no_cache to readdir',
            test: 'node example.js',
            _: [
                ['2m19.183s', { first_run: true }],
                ['2m17.394s']
            ]
        },
        {
            a: 'after fixing an issue in ReadFileCacheFAO',
            test: 'node example.js',
            _: [
                ['3m16.212s', { first_run: true, oops: 'forgot to turn off log output' }],
                ['2m19.675s', { first_run: true }],
                ['2m9.703s'],
                ['2m6.277s'],
            ]
        },
        {
            a: 'experiment with "infinite" directory tree cache ttl',
            test: 'node example.js',
            _: [
                ['2m9.931s', { first_run: true }],
                ['1m58.406s'],
            ]
        },
        {
            a: 'performance of `ls node_modules` with last change',
            test: 'ls node_modules',
            _: [
                ['0m11.233s', { first_run: true }],
                ['0m11.235s', { first_run: true }],
            ]
        },
        {
            a: 'fixing no_thumbs and no_cache',
            test: 'ls node_modules',
            _: [
                ['5.246s', { first_run: true }],
                ['5.090s'],
                ['4.915s'],
                ['5.256s'],
                ['4.858s'],
            ]
        },
        {
            a: 'add missing Path->LocalUID association in readdir',
            test: 'ls node_modules',
            _: [
                ['0.790s', { first_run: true }],
                ['0.731s'],
                ['0.622s'],
                ['0.613s'],
            ]
        },
        {
            a: 'test example express app again',
            test: 'node example.js',
            _: [
                ['14.503s', { first_run: true }],
                ['0.354s'],
                ['0.364s'],
            ]
        }
    ]
})
