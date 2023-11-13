{
    "name": "New blueprint",
    "description": "",
    "version": "",
    "blueprint": {
        "cm": {
            "undo": [
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "c5ffc1d6-2f79-4468-826c-76dba2ae4a33",
                            "type": "nebulant.rectangle.vertical.executionControl.Start",
                            "previous": {
                                "position": {
                                    "x": 2200,
                                    "y": 1240
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 2020,
                                    "y": 1020
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "add",
                        "data": {
                            "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                            "type": "nebulant.rectangle.vertical.generic.UploadFiles",
                            "previous": {},
                            "next": {},
                            "attributes": {
                                "position": {
                                    "x": 1875,
                                    "y": 1295
                                },
                                "size": {
                                    "width": 100,
                                    "height": 120
                                },
                                "angle": 0,
                                "type": "nebulant.rectangle.vertical.generic.UploadFiles",
                                "data": {
                                    "id": "upload-files",
                                    "version": "1.0.2",
                                    "provider": "generic"
                                },
                                "ports": {
                                    "items": [
                                        {
                                            "group": "in",
                                            "attrs": {},
                                            "id": "dfdea216-8adf-40be-953e-3ccc14c24d52"
                                        },
                                        {
                                            "group": "out-ko",
                                            "attrs": {},
                                            "id": "9374a4f0-43dd-469a-a9ba-5d4bcebeea1c"
                                        },
                                        {
                                            "group": "out-ok",
                                            "attrs": {},
                                            "id": "92c8e05b-1d21-4482-9c21-ad970c67b218"
                                        }
                                    ]
                                },
                                "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                                "z": 2
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    },
                    {
                        "action": "change:data",
                        "data": {
                            "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                            "type": "nebulant.rectangle.vertical.generic.UploadFiles",
                            "previous": {
                                "data": {
                                    "id": "upload-files",
                                    "version": "1.0.2",
                                    "provider": "generic"
                                }
                            },
                            "next": {
                                "data": {
                                    "id": "upload-files",
                                    "version": "1.0.2",
                                    "provider": "generic",
                                    "settings": {
                                        "parameters": {
                                            "proxies": [],
                                            "port": 22,
                                            "password": "",
                                            "passphrase": "",
                                            "privkey": "",
                                            "privkeyPath": "",
                                            "username": "",
                                            "target": [],
                                            "_credentials": "privkeyPath",
                                            "paths": [],
                                            "_maxRetries": 5
                                        },
                                        "info": ""
                                    }
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                "data/settings"
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "add",
                        "data": {
                            "id": "55da647d-dbe0-4756-a10a-9af2269cbf11",
                            "type": "nebulant.link.Smart",
                            "previous": {},
                            "next": {},
                            "attributes": {
                                "type": "nebulant.link.Smart",
                                "source": {
                                    "id": "c5ffc1d6-2f79-4468-826c-76dba2ae4a33",
                                    "magnet": "circle",
                                    "port": "69526ce7-29a3-43cd-86e8-ad94c90c63b7"
                                },
                                "target": {
                                    "x": 2080,
                                    "y": 1140
                                },
                                "router": {
                                    "name": "manhattan",
                                    "args": {
                                        "padding": 20
                                    }
                                },
                                "connector": {
                                    "name": "jumpover",
                                    "args": {
                                        "jump": "gap",
                                        "radius": 10
                                    }
                                },
                                "id": "55da647d-dbe0-4756-a10a-9af2269cbf11",
                                "z": 3
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    },
                    {
                        "action": "change:target",
                        "data": {
                            "id": "55da647d-dbe0-4756-a10a-9af2269cbf11",
                            "type": "nebulant.link.Smart",
                            "previous": {
                                "target": {
                                    "x": 2080,
                                    "y": 1140
                                }
                            },
                            "next": {
                                "target": {
                                    "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                                    "magnet": "circle",
                                    "port": "dfdea216-8adf-40be-953e-3ccc14c24d52"
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    },
                    {
                        "action": "change:_not_fully_created",
                        "data": {
                            "id": "55da647d-dbe0-4756-a10a-9af2269cbf11",
                            "type": "nebulant.link.Smart",
                            "previous": {
                                "_not_fully_created": true
                            },
                            "next": {
                                "_not_fully_created": false
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                "_not_fully_created"
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                            "type": "nebulant.rectangle.vertical.generic.UploadFiles",
                            "previous": {
                                "position": {
                                    "x": 1875,
                                    "y": 1295
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 2080,
                                    "y": 1300
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:data",
                        "data": {
                            "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                            "type": "nebulant.rectangle.vertical.generic.UploadFiles",
                            "previous": {
                                "data": {
                                    "id": "upload-files",
                                    "version": "1.0.2",
                                    "provider": "generic",
                                    "settings": {
                                        "parameters": {
                                            "proxies": [],
                                            "port": 22,
                                            "password": "",
                                            "passphrase": "",
                                            "privkey": "",
                                            "privkeyPath": "",
                                            "username": "",
                                            "target": [],
                                            "_credentials": "privkeyPath",
                                            "paths": [],
                                            "_maxRetries": 5
                                        },
                                        "info": ""
                                    }
                                }
                            },
                            "next": {
                                "data": {
                                    "id": "upload-files",
                                    "version": "1.0.2",
                                    "provider": "generic",
                                    "settings": {
                                        "parameters": {
                                            "proxies": [],
                                            "port": 22,
                                            "password": "",
                                            "passphrase": "",
                                            "privkey": "",
                                            "privkeyPath": "",
                                            "username": "",
                                            "target": [],
                                            "_credentials": "privkeyPath",
                                            "paths": [
                                                {
                                                    "_src_type": "file",
                                                    "src": "{{ SOURCEPATH }}",
                                                    "dest": "{{ DSTPATH }}",
                                                    "overwrite": false,
                                                    "recursive": true
                                                }
                                            ],
                                            "_maxRetries": 5
                                        },
                                        "info": ""
                                    }
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                "data/settings"
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                            "type": "nebulant.rectangle.vertical.generic.UploadFiles",
                            "previous": {
                                "position": {
                                    "x": 2080,
                                    "y": 1300
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 1760,
                                    "y": 1440
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "remove",
                        "data": {
                            "id": "55da647d-dbe0-4756-a10a-9af2269cbf11",
                            "type": "nebulant.link.Smart",
                            "previous": {},
                            "next": {},
                            "attributes": {
                                "type": "nebulant.link.Smart",
                                "source": {
                                    "id": "c5ffc1d6-2f79-4468-826c-76dba2ae4a33",
                                    "magnet": "circle",
                                    "port": "69526ce7-29a3-43cd-86e8-ad94c90c63b7"
                                },
                                "target": {
                                    "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                                    "magnet": "circle",
                                    "port": "dfdea216-8adf-40be-953e-3ccc14c24d52"
                                },
                                "router": {
                                    "name": "manhattan",
                                    "args": {
                                        "padding": 20
                                    }
                                },
                                "connector": {
                                    "name": "jumpover",
                                    "args": {
                                        "jump": "gap",
                                        "radius": 10
                                    }
                                },
                                "id": "55da647d-dbe0-4756-a10a-9af2269cbf11",
                                "z": 3
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "add",
                        "data": {
                            "id": "c1c4f5d6-c8d0-4305-981a-045eb988682a",
                            "type": "nebulant.rectangle.vertical.generic.WriteFile",
                            "previous": {},
                            "next": {},
                            "attributes": {
                                "position": {
                                    "x": 2115,
                                    "y": 1175
                                },
                                "size": {
                                    "width": 100,
                                    "height": 120
                                },
                                "angle": 0,
                                "type": "nebulant.rectangle.vertical.generic.WriteFile",
                                "data": {
                                    "id": "write-file",
                                    "version": "1.0.0",
                                    "provider": "generic"
                                },
                                "ports": {
                                    "items": [
                                        {
                                            "group": "in",
                                            "attrs": {},
                                            "id": "5705957e-e0b8-4e9f-b453-e8d021e44d53"
                                        },
                                        {
                                            "group": "out-ko",
                                            "attrs": {},
                                            "id": "b21cdeaf-d269-4c32-9258-aae0b74b7056"
                                        },
                                        {
                                            "group": "out-ok",
                                            "attrs": {},
                                            "id": "c12d4822-68f9-4d78-aed8-b7fc256f07ed"
                                        }
                                    ]
                                },
                                "id": "c1c4f5d6-c8d0-4305-981a-045eb988682a",
                                "z": 5
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    },
                    {
                        "action": "change:data",
                        "data": {
                            "id": "c1c4f5d6-c8d0-4305-981a-045eb988682a",
                            "type": "nebulant.rectangle.vertical.generic.WriteFile",
                            "previous": {
                                "data": {
                                    "id": "write-file",
                                    "version": "1.0.0",
                                    "provider": "generic"
                                }
                            },
                            "next": {
                                "data": {
                                    "id": "write-file",
                                    "version": "1.0.0",
                                    "provider": "generic",
                                    "settings": {
                                        "outputs": {
                                            "result": {
                                                "hasID": false,
                                                "waiters": [],
                                                "async": false,
                                                "capabilities": [],
                                                "provider": "generic",
                                                "type": "file_io",
                                                "value": "WRITE_OPERATION"
                                            }
                                        },
                                        "parameters": {
                                            "interpolate": true,
                                            "content": "",
                                            "file_path": "",
                                            "_maxRetries": 5
                                        },
                                        "info": ""
                                    }
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                "data/settings"
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:data",
                        "data": {
                            "id": "c1c4f5d6-c8d0-4305-981a-045eb988682a",
                            "type": "nebulant.rectangle.vertical.generic.WriteFile",
                            "previous": {
                                "data": {
                                    "id": "write-file",
                                    "version": "1.0.0",
                                    "provider": "generic",
                                    "settings": {
                                        "outputs": {
                                            "result": {
                                                "hasID": false,
                                                "waiters": [],
                                                "async": false,
                                                "capabilities": [],
                                                "provider": "generic",
                                                "type": "file_io",
                                                "value": "WRITE_OPERATION"
                                            }
                                        },
                                        "parameters": {
                                            "interpolate": true,
                                            "content": "",
                                            "file_path": "",
                                            "_maxRetries": 5
                                        },
                                        "info": ""
                                    }
                                }
                            },
                            "next": {
                                "data": {
                                    "id": "write-file",
                                    "version": "1.0.0",
                                    "provider": "generic",
                                    "settings": {
                                        "outputs": {
                                            "result": {
                                                "hasID": false,
                                                "waiters": [],
                                                "async": false,
                                                "capabilities": [],
                                                "provider": "generic",
                                                "type": "file_io",
                                                "value": "WRITE_OPERATION"
                                            }
                                        },
                                        "parameters": {
                                            "interpolate": true,
                                            "content": "{\n    \"versions\": {\n        \"latest\": {\n            \"version\": \"{{ VERSION_NUMBER }}\",\n            \"date\": \"{{ VERSION_DATE }}\",\n            \"url\": \"https://cli-releases/1.0.1/nebulant-{OS}-{ARCH}{EXE}\",\n            \"checksum\": \"{URL}.checksum\"\n        }\n    }\n}\n",
                                            "file_path": "./dist/version.json",
                                            "_maxRetries": 5
                                        },
                                        "info": ""
                                    }
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                "data/settings"
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "c1c4f5d6-c8d0-4305-981a-045eb988682a",
                            "type": "nebulant.rectangle.vertical.generic.WriteFile",
                            "previous": {
                                "position": {
                                    "x": 2115,
                                    "y": 1175
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 2120,
                                    "y": 1260
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "c1c4f5d6-c8d0-4305-981a-045eb988682a",
                            "type": "nebulant.rectangle.vertical.generic.WriteFile",
                            "previous": {
                                "position": {
                                    "x": 2120,
                                    "y": 1260
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 2020,
                                    "y": 1240
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "add",
                        "data": {
                            "id": "19a9ad00-d0ee-444a-8f5d-81b3414236fe",
                            "type": "nebulant.link.Smart",
                            "previous": {},
                            "next": {},
                            "attributes": {
                                "type": "nebulant.link.Smart",
                                "source": {
                                    "id": "c5ffc1d6-2f79-4468-826c-76dba2ae4a33",
                                    "magnet": "circle",
                                    "port": "69526ce7-29a3-43cd-86e8-ad94c90c63b7"
                                },
                                "target": {
                                    "x": 2080,
                                    "y": 1140
                                },
                                "router": {
                                    "name": "manhattan",
                                    "args": {
                                        "padding": 20
                                    }
                                },
                                "connector": {
                                    "name": "jumpover",
                                    "args": {
                                        "jump": "gap",
                                        "radius": 10
                                    }
                                },
                                "id": "19a9ad00-d0ee-444a-8f5d-81b3414236fe",
                                "z": 6
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    },
                    {
                        "action": "change:target",
                        "data": {
                            "id": "19a9ad00-d0ee-444a-8f5d-81b3414236fe",
                            "type": "nebulant.link.Smart",
                            "previous": {
                                "target": {
                                    "x": 2080,
                                    "y": 1140
                                }
                            },
                            "next": {
                                "target": {
                                    "id": "c1c4f5d6-c8d0-4305-981a-045eb988682a",
                                    "magnet": "circle",
                                    "port": "5705957e-e0b8-4e9f-b453-e8d021e44d53"
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    },
                    {
                        "action": "change:_not_fully_created",
                        "data": {
                            "id": "19a9ad00-d0ee-444a-8f5d-81b3414236fe",
                            "type": "nebulant.link.Smart",
                            "previous": {
                                "_not_fully_created": true
                            },
                            "next": {
                                "_not_fully_created": false
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                "_not_fully_created"
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "add",
                        "data": {
                            "id": "9de8a0c8-0fa5-4a5b-bdfb-ca2e1a1f841e",
                            "type": "nebulant.rectangle.vertical.generic.ReadFile",
                            "previous": {},
                            "next": {},
                            "attributes": {
                                "position": {
                                    "x": 2155,
                                    "y": 1455
                                },
                                "size": {
                                    "width": 100,
                                    "height": 120
                                },
                                "angle": 0,
                                "type": "nebulant.rectangle.vertical.generic.ReadFile",
                                "data": {
                                    "id": "read-file",
                                    "version": "1.0.0",
                                    "provider": "generic"
                                },
                                "ports": {
                                    "items": [
                                        {
                                            "group": "in",
                                            "attrs": {},
                                            "id": "87463721-4282-4982-872c-f7cc6efc71b5"
                                        },
                                        {
                                            "group": "out-ko",
                                            "attrs": {},
                                            "id": "e87af4a3-879a-4a00-99a2-96c2c192bba0"
                                        },
                                        {
                                            "group": "out-ok",
                                            "attrs": {},
                                            "id": "9aa44407-5369-4a85-b789-44b2fb38dde0"
                                        }
                                    ]
                                },
                                "id": "9de8a0c8-0fa5-4a5b-bdfb-ca2e1a1f841e",
                                "z": 8
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    },
                    {
                        "action": "change:data",
                        "data": {
                            "id": "9de8a0c8-0fa5-4a5b-bdfb-ca2e1a1f841e",
                            "type": "nebulant.rectangle.vertical.generic.ReadFile",
                            "previous": {
                                "data": {
                                    "id": "read-file",
                                    "version": "1.0.0",
                                    "provider": "generic"
                                }
                            },
                            "next": {
                                "data": {
                                    "id": "read-file",
                                    "version": "1.0.0",
                                    "provider": "generic",
                                    "settings": {
                                        "outputs": {
                                            "result": {
                                                "hasID": false,
                                                "waiters": [],
                                                "async": false,
                                                "capabilities": [],
                                                "provider": "generic",
                                                "type": "user variable",
                                                "value": "FILE_CONTENT"
                                            }
                                        },
                                        "parameters": {
                                            "interpolate": true,
                                            "file_path": ""
                                        },
                                        "info": ""
                                    }
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                "data/settings"
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "add",
                        "data": {
                            "id": "05295225-025d-41df-b263-b53c404c907c",
                            "type": "nebulant.link.Smart",
                            "previous": {},
                            "next": {},
                            "attributes": {
                                "type": "nebulant.link.Smart",
                                "source": {
                                    "id": "c1c4f5d6-c8d0-4305-981a-045eb988682a",
                                    "magnet": "circle",
                                    "port": "c12d4822-68f9-4d78-aed8-b7fc256f07ed"
                                },
                                "target": {
                                    "x": 2120,
                                    "y": 1360
                                },
                                "router": {
                                    "name": "manhattan",
                                    "args": {
                                        "padding": 20
                                    }
                                },
                                "connector": {
                                    "name": "jumpover",
                                    "args": {
                                        "jump": "gap",
                                        "radius": 10
                                    }
                                },
                                "id": "05295225-025d-41df-b263-b53c404c907c",
                                "z": 9
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    },
                    {
                        "action": "change:target",
                        "data": {
                            "id": "05295225-025d-41df-b263-b53c404c907c",
                            "type": "nebulant.link.Smart",
                            "previous": {
                                "target": {
                                    "x": 2120,
                                    "y": 1360
                                }
                            },
                            "next": {
                                "target": {
                                    "id": "9de8a0c8-0fa5-4a5b-bdfb-ca2e1a1f841e",
                                    "magnet": "circle",
                                    "port": "87463721-4282-4982-872c-f7cc6efc71b5"
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    },
                    {
                        "action": "change:_not_fully_created",
                        "data": {
                            "id": "05295225-025d-41df-b263-b53c404c907c",
                            "type": "nebulant.link.Smart",
                            "previous": {
                                "_not_fully_created": true
                            },
                            "next": {
                                "_not_fully_created": false
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                "_not_fully_created"
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:data",
                        "data": {
                            "id": "9de8a0c8-0fa5-4a5b-bdfb-ca2e1a1f841e",
                            "type": "nebulant.rectangle.vertical.generic.ReadFile",
                            "previous": {
                                "data": {
                                    "id": "read-file",
                                    "version": "1.0.0",
                                    "provider": "generic",
                                    "settings": {
                                        "outputs": {
                                            "result": {
                                                "hasID": false,
                                                "waiters": [],
                                                "async": false,
                                                "capabilities": [],
                                                "provider": "generic",
                                                "type": "user variable",
                                                "value": "FILE_CONTENT"
                                            }
                                        },
                                        "parameters": {
                                            "interpolate": true,
                                            "file_path": ""
                                        },
                                        "info": ""
                                    }
                                }
                            },
                            "next": {
                                "data": {
                                    "id": "read-file",
                                    "version": "1.0.0",
                                    "provider": "generic",
                                    "settings": {
                                        "outputs": {
                                            "result": {
                                                "hasID": false,
                                                "waiters": [],
                                                "async": false,
                                                "capabilities": [],
                                                "provider": "generic",
                                                "type": "user variable",
                                                "value": "VERSION_JSON_CONTENT"
                                            }
                                        },
                                        "parameters": {
                                            "interpolate": true,
                                            "file_path": "./dist/version.json"
                                        },
                                        "info": ""
                                    }
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                "data/settings"
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "c1c4f5d6-c8d0-4305-981a-045eb988682a",
                            "type": "nebulant.rectangle.vertical.generic.WriteFile",
                            "previous": {
                                "position": {
                                    "x": 2020,
                                    "y": 1240
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 2020,
                                    "y": 1180
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "9de8a0c8-0fa5-4a5b-bdfb-ca2e1a1f841e",
                            "type": "nebulant.rectangle.vertical.generic.ReadFile",
                            "previous": {
                                "position": {
                                    "x": 2155,
                                    "y": 1455
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 2020,
                                    "y": 1340
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "add",
                        "data": {
                            "id": "0b3e6457-558c-43fa-a2be-525d1213aa47",
                            "type": "nebulant.rectangle.vertical.generic.Log",
                            "previous": {},
                            "next": {},
                            "attributes": {
                                "position": {
                                    "x": 2215,
                                    "y": 1335
                                },
                                "size": {
                                    "width": 100,
                                    "height": 120
                                },
                                "angle": 0,
                                "type": "nebulant.rectangle.vertical.generic.Log",
                                "data": {
                                    "id": "log",
                                    "version": "1.0.0",
                                    "provider": "generic"
                                },
                                "ports": {
                                    "items": [
                                        {
                                            "group": "in",
                                            "attrs": {},
                                            "id": "ede9d391-e1e0-4ad5-9126-85c6174e73ba"
                                        },
                                        {
                                            "group": "out-ko",
                                            "attrs": {},
                                            "id": "e57fda9f-fcbe-4761-b09b-22c3bfbe397d"
                                        },
                                        {
                                            "group": "out-ok",
                                            "attrs": {},
                                            "id": "63e1e210-ffac-43b5-ae3b-5e86d2d25b7c"
                                        }
                                    ]
                                },
                                "id": "0b3e6457-558c-43fa-a2be-525d1213aa47",
                                "z": 13
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    },
                    {
                        "action": "change:data",
                        "data": {
                            "id": "0b3e6457-558c-43fa-a2be-525d1213aa47",
                            "type": "nebulant.rectangle.vertical.generic.Log",
                            "previous": {
                                "data": {
                                    "id": "log",
                                    "version": "1.0.0",
                                    "provider": "generic"
                                }
                            },
                            "next": {
                                "data": {
                                    "id": "log",
                                    "version": "1.0.0",
                                    "provider": "generic",
                                    "settings": {
                                        "parameters": {
                                            "content": ""
                                        },
                                        "info": ""
                                    }
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                "data/settings"
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "add",
                        "data": {
                            "id": "649ae3e3-bea3-4aa0-a3ca-6444b0631cfe",
                            "type": "nebulant.link.Smart",
                            "previous": {},
                            "next": {},
                            "attributes": {
                                "type": "nebulant.link.Smart",
                                "source": {
                                    "id": "9de8a0c8-0fa5-4a5b-bdfb-ca2e1a1f841e",
                                    "magnet": "circle",
                                    "port": "9aa44407-5369-4a85-b789-44b2fb38dde0"
                                },
                                "target": {
                                    "x": 2120,
                                    "y": 1460
                                },
                                "router": {
                                    "name": "manhattan",
                                    "args": {
                                        "padding": 20
                                    }
                                },
                                "connector": {
                                    "name": "jumpover",
                                    "args": {
                                        "jump": "gap",
                                        "radius": 10
                                    }
                                },
                                "id": "649ae3e3-bea3-4aa0-a3ca-6444b0631cfe",
                                "z": 14
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    },
                    {
                        "action": "change:target",
                        "data": {
                            "id": "649ae3e3-bea3-4aa0-a3ca-6444b0631cfe",
                            "type": "nebulant.link.Smart",
                            "previous": {
                                "target": {
                                    "x": 2120,
                                    "y": 1460
                                }
                            },
                            "next": {
                                "target": {
                                    "id": "0b3e6457-558c-43fa-a2be-525d1213aa47",
                                    "magnet": "circle",
                                    "port": "ede9d391-e1e0-4ad5-9126-85c6174e73ba"
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    },
                    {
                        "action": "change:_not_fully_created",
                        "data": {
                            "id": "649ae3e3-bea3-4aa0-a3ca-6444b0631cfe",
                            "type": "nebulant.link.Smart",
                            "previous": {
                                "_not_fully_created": true
                            },
                            "next": {
                                "_not_fully_created": false
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                "_not_fully_created"
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:data",
                        "data": {
                            "id": "0b3e6457-558c-43fa-a2be-525d1213aa47",
                            "type": "nebulant.rectangle.vertical.generic.Log",
                            "previous": {
                                "data": {
                                    "id": "log",
                                    "version": "1.0.0",
                                    "provider": "generic",
                                    "settings": {
                                        "parameters": {
                                            "content": ""
                                        },
                                        "info": ""
                                    }
                                }
                            },
                            "next": {
                                "data": {
                                    "id": "log",
                                    "version": "1.0.0",
                                    "provider": "generic",
                                    "settings": {
                                        "parameters": {
                                            "content": "GENERATED version.json content:\n\n{{ VERSION_JSON_CONTENT }}"
                                        },
                                        "info": ""
                                    }
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                "data/settings"
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                            "type": "nebulant.rectangle.vertical.generic.UploadFiles",
                            "previous": {
                                "position": {
                                    "x": 1760,
                                    "y": 1440
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 1940,
                                    "y": 1540
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "add",
                        "data": {
                            "id": "f8860712-757e-4bc7-9267-e6e0ff9db287",
                            "type": "nebulant.link.Smart",
                            "previous": {},
                            "next": {},
                            "attributes": {
                                "type": "nebulant.link.Smart",
                                "source": {
                                    "id": "0b3e6457-558c-43fa-a2be-525d1213aa47",
                                    "magnet": "circle",
                                    "port": "63e1e210-ffac-43b5-ae3b-5e86d2d25b7c"
                                },
                                "target": {
                                    "x": 2320,
                                    "y": 1460
                                },
                                "router": {
                                    "name": "manhattan",
                                    "args": {
                                        "padding": 20
                                    }
                                },
                                "connector": {
                                    "name": "jumpover",
                                    "args": {
                                        "jump": "gap",
                                        "radius": 10
                                    }
                                },
                                "id": "f8860712-757e-4bc7-9267-e6e0ff9db287",
                                "z": 17
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    },
                    {
                        "action": "change:target",
                        "data": {
                            "id": "f8860712-757e-4bc7-9267-e6e0ff9db287",
                            "type": "nebulant.link.Smart",
                            "previous": {
                                "target": {
                                    "x": 2320,
                                    "y": 1460
                                }
                            },
                            "next": {
                                "target": {
                                    "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                                    "magnet": "circle",
                                    "port": "dfdea216-8adf-40be-953e-3ccc14c24d52"
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    },
                    {
                        "action": "change:_not_fully_created",
                        "data": {
                            "id": "f8860712-757e-4bc7-9267-e6e0ff9db287",
                            "type": "nebulant.link.Smart",
                            "previous": {
                                "_not_fully_created": true
                            },
                            "next": {
                                "_not_fully_created": false
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                "_not_fully_created"
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                            "type": "nebulant.rectangle.vertical.generic.UploadFiles",
                            "previous": {
                                "position": {
                                    "x": 1940,
                                    "y": 1540
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 2220,
                                    "y": 1500
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "9de8a0c8-0fa5-4a5b-bdfb-ca2e1a1f841e",
                            "type": "nebulant.rectangle.vertical.generic.ReadFile",
                            "previous": {
                                "position": {
                                    "x": 2020,
                                    "y": 1340
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 2160,
                                    "y": 1200
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "0b3e6457-558c-43fa-a2be-525d1213aa47",
                            "type": "nebulant.rectangle.vertical.generic.Log",
                            "previous": {
                                "position": {
                                    "x": 2215,
                                    "y": 1335
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 1960,
                                    "y": 1600
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                            "type": "nebulant.rectangle.vertical.generic.UploadFiles",
                            "previous": {
                                "position": {
                                    "x": 2220,
                                    "y": 1500
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 2360,
                                    "y": 1580
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "9de8a0c8-0fa5-4a5b-bdfb-ca2e1a1f841e",
                            "type": "nebulant.rectangle.vertical.generic.ReadFile",
                            "previous": {
                                "position": {
                                    "x": 2160,
                                    "y": 1200
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 2160,
                                    "y": 1180
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:vertices",
                        "data": {
                            "id": "05295225-025d-41df-b263-b53c404c907c",
                            "type": "nebulant.link.Smart",
                            "previous": {},
                            "next": {
                                "vertices": [
                                    {
                                        "x": 2140,
                                        "y": 1160
                                    }
                                ]
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "0b3e6457-558c-43fa-a2be-525d1213aa47",
                            "type": "nebulant.rectangle.vertical.generic.Log",
                            "previous": {
                                "position": {
                                    "x": 1960,
                                    "y": 1600
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 2160,
                                    "y": 1340
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                            "type": "nebulant.rectangle.vertical.generic.UploadFiles",
                            "previous": {
                                "position": {
                                    "x": 2360,
                                    "y": 1580
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 2160,
                                    "y": 1500
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "0b3e6457-558c-43fa-a2be-525d1213aa47",
                            "type": "nebulant.rectangle.vertical.generic.Log",
                            "previous": {
                                "position": {
                                    "x": 2160,
                                    "y": 1340
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 2280,
                                    "y": 1180
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:vertices",
                        "data": {
                            "id": "649ae3e3-bea3-4aa0-a3ca-6444b0631cfe",
                            "type": "nebulant.link.Smart",
                            "previous": {},
                            "next": {
                                "vertices": [
                                    {
                                        "x": 2300,
                                        "y": 1160
                                    }
                                ]
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                            "type": "nebulant.rectangle.vertical.generic.UploadFiles",
                            "previous": {
                                "position": {
                                    "x": 2160,
                                    "y": 1500
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 2160,
                                    "y": 1380
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:position",
                        "data": {
                            "id": "c5ffc1d6-2f79-4468-826c-76dba2ae4a33",
                            "type": "nebulant.rectangle.vertical.executionControl.Start",
                            "previous": {
                                "position": {
                                    "x": 2020,
                                    "y": 1020
                                }
                            },
                            "next": {
                                "position": {
                                    "x": 2020,
                                    "y": 960
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                null
                            ]
                        }
                    }
                ],
                [
                    {
                        "action": "change:data",
                        "data": {
                            "id": "c1c4f5d6-c8d0-4305-981a-045eb988682a",
                            "type": "nebulant.rectangle.vertical.generic.WriteFile",
                            "previous": {
                                "data": {
                                    "id": "write-file",
                                    "version": "1.0.0",
                                    "provider": "generic",
                                    "settings": {
                                        "outputs": {
                                            "result": {
                                                "hasID": false,
                                                "waiters": [],
                                                "async": false,
                                                "capabilities": [],
                                                "provider": "generic",
                                                "type": "file_io",
                                                "value": "WRITE_OPERATION"
                                            }
                                        },
                                        "parameters": {
                                            "interpolate": true,
                                            "content": "{\n    \"versions\": {\n        \"latest\": {\n            \"version\": \"{{ VERSION_NUMBER }}\",\n            \"date\": \"{{ VERSION_DATE }}\",\n            \"url\": \"https://cli-releases/1.0.1/nebulant-{OS}-{ARCH}{EXE}\",\n            \"checksum\": \"{URL}.checksum\"\n        }\n    }\n}\n",
                                            "file_path": "./dist/version.json",
                                            "_maxRetries": 5
                                        },
                                        "info": ""
                                    }
                                }
                            },
                            "next": {
                                "data": {
                                    "id": "write-file",
                                    "version": "1.0.0",
                                    "provider": "generic",
                                    "settings": {
                                        "outputs": {
                                            "result": {
                                                "hasID": false,
                                                "waiters": [],
                                                "async": false,
                                                "capabilities": [],
                                                "provider": "generic",
                                                "type": "file_io",
                                                "value": "WRITE_OPERATION"
                                            }
                                        },
                                        "parameters": {
                                            "interpolate": true,
                                            "content": "{\n    \"versions\": {\n        \"latest\": {\n            \"version\": \"{{ VERSION_NUMBER }}\",\n            \"date\": \"{{ VERSION_DATE }}\",\n            \"url\": \"https://{{ URL_DOMAIN }}/1.0.1/nebulant-{OS}-{ARCH}{EXE}\",\n            \"checksum\": \"{URL}.checksum\"\n        }\n    }\n}\n",
                                            "file_path": "./dist/version.json",
                                            "_maxRetries": 5
                                        },
                                        "info": ""
                                    }
                                }
                            }
                        },
                        "batch": true,
                        "options": {
                            "propertyPath": [
                                "data/settings"
                            ]
                        }
                    }
                ]
            ],
            "redo": []
        },
        "diagram": {
            "cells": [
                {
                    "type": "nebulant.link.Smart",
                    "source": {
                        "id": "c5ffc1d6-2f79-4468-826c-76dba2ae4a33",
                        "magnet": "circle",
                        "port": "69526ce7-29a3-43cd-86e8-ad94c90c63b7"
                    },
                    "target": {
                        "id": "c1c4f5d6-c8d0-4305-981a-045eb988682a",
                        "magnet": "circle",
                        "port": "5705957e-e0b8-4e9f-b453-e8d021e44d53"
                    },
                    "router": {
                        "name": "manhattan",
                        "args": {
                            "padding": 20
                        }
                    },
                    "connector": {
                        "name": "jumpover",
                        "args": {
                            "jump": "gap",
                            "radius": 10
                        }
                    },
                    "id": "19a9ad00-d0ee-444a-8f5d-81b3414236fe",
                    "z": 6
                },
                {
                    "type": "nebulant.link.Smart",
                    "source": {
                        "id": "c1c4f5d6-c8d0-4305-981a-045eb988682a",
                        "magnet": "circle",
                        "port": "c12d4822-68f9-4d78-aed8-b7fc256f07ed"
                    },
                    "target": {
                        "id": "9de8a0c8-0fa5-4a5b-bdfb-ca2e1a1f841e",
                        "magnet": "circle",
                        "port": "87463721-4282-4982-872c-f7cc6efc71b5"
                    },
                    "router": {
                        "name": "manhattan",
                        "args": {
                            "padding": 20
                        }
                    },
                    "connector": {
                        "name": "jumpover",
                        "args": {
                            "jump": "gap",
                            "radius": 10
                        }
                    },
                    "id": "05295225-025d-41df-b263-b53c404c907c",
                    "z": 9
                },
                {
                    "type": "nebulant.link.Smart",
                    "source": {
                        "id": "9de8a0c8-0fa5-4a5b-bdfb-ca2e1a1f841e",
                        "magnet": "circle",
                        "port": "9aa44407-5369-4a85-b789-44b2fb38dde0"
                    },
                    "target": {
                        "id": "0b3e6457-558c-43fa-a2be-525d1213aa47",
                        "magnet": "circle",
                        "port": "ede9d391-e1e0-4ad5-9126-85c6174e73ba"
                    },
                    "router": {
                        "name": "manhattan",
                        "args": {
                            "padding": 20
                        }
                    },
                    "connector": {
                        "name": "jumpover",
                        "args": {
                            "jump": "gap",
                            "radius": 10
                        }
                    },
                    "id": "649ae3e3-bea3-4aa0-a3ca-6444b0631cfe",
                    "z": 14
                },
                {
                    "type": "nebulant.link.Smart",
                    "source": {
                        "id": "0b3e6457-558c-43fa-a2be-525d1213aa47",
                        "magnet": "circle",
                        "port": "63e1e210-ffac-43b5-ae3b-5e86d2d25b7c"
                    },
                    "target": {
                        "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                        "magnet": "circle",
                        "port": "dfdea216-8adf-40be-953e-3ccc14c24d52"
                    },
                    "router": {
                        "name": "manhattan",
                        "args": {
                            "padding": 20
                        }
                    },
                    "connector": {
                        "name": "jumpover",
                        "args": {
                            "jump": "gap",
                            "radius": 10
                        }
                    },
                    "id": "f8860712-757e-4bc7-9267-e6e0ff9db287",
                    "z": 17
                },
                {
                    "position": {
                        "x": 2160,
                        "y": 1180
                    },
                    "size": {
                        "width": 100,
                        "height": 120
                    },
                    "angle": 0,
                    "type": "nebulant.rectangle.vertical.generic.ReadFile",
                    "data": {
                        "id": "read-file",
                        "version": "1.0.0",
                        "provider": "generic",
                        "settings": {
                            "outputs": {
                                "result": {
                                    "hasID": false,
                                    "waiters": [],
                                    "async": false,
                                    "capabilities": [],
                                    "provider": "generic",
                                    "type": "user variable",
                                    "value": "VERSION_JSON_CONTENT"
                                }
                            },
                            "parameters": {
                                "interpolate": true,
                                "file_path": "./dist/version.json"
                            },
                            "info": ""
                        }
                    },
                    "ports": {
                        "items": [
                            {
                                "group": "in",
                                "attrs": {},
                                "id": "87463721-4282-4982-872c-f7cc6efc71b5"
                            },
                            {
                                "group": "out-ko",
                                "attrs": {},
                                "id": "e87af4a3-879a-4a00-99a2-96c2c192bba0"
                            },
                            {
                                "group": "out-ok",
                                "attrs": {},
                                "id": "9aa44407-5369-4a85-b789-44b2fb38dde0"
                            }
                        ]
                    },
                    "id": "9de8a0c8-0fa5-4a5b-bdfb-ca2e1a1f841e",
                    "z": 22
                },
                {
                    "position": {
                        "x": 2280,
                        "y": 1180
                    },
                    "size": {
                        "width": 100,
                        "height": 120
                    },
                    "angle": 0,
                    "type": "nebulant.rectangle.vertical.generic.Log",
                    "data": {
                        "id": "log",
                        "version": "1.0.0",
                        "provider": "generic",
                        "settings": {
                            "parameters": {
                                "content": "GENERATED version.json content:\n\n{{ VERSION_JSON_CONTENT }}"
                            },
                            "info": ""
                        }
                    },
                    "ports": {
                        "items": [
                            {
                                "group": "in",
                                "attrs": {},
                                "id": "ede9d391-e1e0-4ad5-9126-85c6174e73ba"
                            },
                            {
                                "group": "out-ko",
                                "attrs": {},
                                "id": "e57fda9f-fcbe-4761-b09b-22c3bfbe397d"
                            },
                            {
                                "group": "out-ok",
                                "attrs": {},
                                "id": "63e1e210-ffac-43b5-ae3b-5e86d2d25b7c"
                            }
                        ]
                    },
                    "id": "0b3e6457-558c-43fa-a2be-525d1213aa47",
                    "z": 25
                },
                {
                    "position": {
                        "x": 2160,
                        "y": 1380
                    },
                    "size": {
                        "width": 100,
                        "height": 120
                    },
                    "angle": 0,
                    "type": "nebulant.rectangle.vertical.generic.UploadFiles",
                    "data": {
                        "id": "upload-files",
                        "version": "1.0.2",
                        "provider": "generic",
                        "settings": {
                            "parameters": {
                                "proxies": [],
                                "port": 22,
                                "password": "",
                                "passphrase": "",
                                "privkey": "",
                                "privkeyPath": "",
                                "username": "",
                                "target": [],
                                "_credentials": "privkeyPath",
                                "paths": [
                                    {
                                        "_src_type": "file",
                                        "src": "{{ SOURCEPATH }}",
                                        "dest": "{{ DSTPATH }}",
                                        "overwrite": false,
                                        "recursive": true
                                    }
                                ],
                                "_maxRetries": 5
                            },
                            "info": ""
                        }
                    },
                    "ports": {
                        "items": [
                            {
                                "group": "in",
                                "attrs": {},
                                "id": "dfdea216-8adf-40be-953e-3ccc14c24d52"
                            },
                            {
                                "group": "out-ko",
                                "attrs": {},
                                "id": "9374a4f0-43dd-469a-a9ba-5d4bcebeea1c"
                            },
                            {
                                "group": "out-ok",
                                "attrs": {},
                                "id": "92c8e05b-1d21-4482-9c21-ad970c67b218"
                            }
                        ]
                    },
                    "id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                    "z": 26
                },
                {
                    "position": {
                        "x": 2020,
                        "y": 960
                    },
                    "size": {
                        "width": 100,
                        "height": 120
                    },
                    "angle": 0,
                    "type": "nebulant.rectangle.vertical.executionControl.Start",
                    "data": {
                        "id": "start",
                        "version": "1.0.10",
                        "provider": "execution-control",
                        "settings": {
                            "parameters": {
                                "input_parameters": [],
                                "image": "",
                                "description": "",
                                "color": "#7986cb",
                                "text_color": "#000000",
                                "version": "draft",
                                "name": "Group",
                                "group_settings_enabled": false
                            },
                            "info": "This is the start node!"
                        }
                    },
                    "ports": {
                        "items": [
                            {
                                "group": "in",
                                "attrs": {},
                                "id": "1ac83cfa-0e3d-48e0-b68b-154ebd40286f"
                            },
                            {
                                "group": "out-ok",
                                "attrs": {},
                                "id": "69526ce7-29a3-43cd-86e8-ad94c90c63b7"
                            }
                        ]
                    },
                    "id": "c5ffc1d6-2f79-4468-826c-76dba2ae4a33",
                    "z": 27
                },
                {
                    "position": {
                        "x": 2020,
                        "y": 1180
                    },
                    "size": {
                        "width": 100,
                        "height": 120
                    },
                    "angle": 0,
                    "type": "nebulant.rectangle.vertical.generic.WriteFile",
                    "data": {
                        "id": "write-file",
                        "version": "1.0.0",
                        "provider": "generic",
                        "settings": {
                            "outputs": {
                                "result": {
                                    "hasID": false,
                                    "waiters": [],
                                    "async": false,
                                    "capabilities": [],
                                    "provider": "generic",
                                    "type": "file_io",
                                    "value": "WRITE_OPERATION"
                                }
                            },
                            "parameters": {
                                "interpolate": true,
                                "content": "{\n    \"versions\": {\n        \"latest\": {\n            \"version\": \"{{ VERSION_NUMBER }}\",\n            \"date\": \"{{ VERSION_DATE }}\",\n            \"url\": \"https://{{ URL_DOMAIN }}/1.0.1/nebulant-{OS}-{ARCH}{EXE}\",\n            \"checksum\": \"{URL}.checksum\"\n        }\n    }\n}\n",
                                "file_path": "./dist/version.json",
                                "_maxRetries": 5
                            },
                            "info": ""
                        }
                    },
                    "ports": {
                        "items": [
                            {
                                "group": "in",
                                "attrs": {},
                                "id": "5705957e-e0b8-4e9f-b453-e8d021e44d53"
                            },
                            {
                                "group": "out-ko",
                                "attrs": {},
                                "id": "b21cdeaf-d269-4c32-9258-aae0b74b7056"
                            },
                            {
                                "group": "out-ok",
                                "attrs": {},
                                "id": "c12d4822-68f9-4d78-aed8-b7fc256f07ed"
                            }
                        ]
                    },
                    "id": "c1c4f5d6-c8d0-4305-981a-045eb988682a",
                    "z": 28
                }
            ]
        },
        "diagram_version": "1.0.2",
        "n_warnings": 0,
        "n_errors": 0,
        "actions": [
            {
                "action_id": "9de8a0c8-0fa5-4a5b-bdfb-ca2e1a1f841e",
                "provider": "generic",
                "version": "1.0.0",
                "action": "read_file",
                "parameters": {
                    "interpolate": true,
                    "file_path": "./dist/version.json"
                },
                "output": "VERSION_JSON_CONTENT",
                "next_action": {
                    "ok": [
                        "0b3e6457-558c-43fa-a2be-525d1213aa47"
                    ]
                },
                "debug_network": false
            },
            {
                "action_id": "0b3e6457-558c-43fa-a2be-525d1213aa47",
                "provider": "generic",
                "version": "1.0.0",
                "action": "log",
                "parameters": {
                    "content": "GENERATED version.json content:\n\n{{ VERSION_JSON_CONTENT }}"
                },
                "next_action": {
                    "ok": [
                        "9f452cdf-cf38-4c86-b536-7614b4898d4c"
                    ]
                },
                "debug_network": false
            },
            {
                "action_id": "9f452cdf-cf38-4c86-b536-7614b4898d4c",
                "provider": "cloudflare",
                "version": "1.0.2",
                "action": "r2_upload",
                "parameters": {
                    "paths": [
                        {
                            "src": "{{ SOURCEPATH }}",
                            "dest": "{{ DSTPATH }}",
                            "bucket": "{{ BUCKET }}"
                        }
                    ]
                },
                "next_action": {},
                "debug_network": false
            },
            {
                "action_id": "c5ffc1d6-2f79-4468-826c-76dba2ae4a33",
                "provider": "generic",
                "version": "1.0.10",
                "first_action": true,
                "action": "start",
                "next_action": {
                    "ok": [
                        "c1c4f5d6-c8d0-4305-981a-045eb988682a"
                    ]
                },
                "debug_network": false
            },
            {
                "action_id": "c1c4f5d6-c8d0-4305-981a-045eb988682a",
                "provider": "generic",
                "version": "1.0.0",
                "action": "write_file",
                "parameters": {
                    "interpolate": true,
                    "content": "{\n    \"versions\": {\n        \"latest\": {\n            \"version\": \"{{ VERSION_NUMBER }}\",\n            \"date\": \"{{ VERSION_DATE }}\",\n            \"url\": \"https://{{ URL_DOMAIN }}/1.0.1/nebulant-{OS}-{ARCH}{EXE}\",\n            \"checksum\": \"{URL}.checksum\"\n        }\n    }\n}\n",
                    "file_path": "./dist/version.json",
                    "_maxRetries": 5
                },
                "output": "WRITE_OPERATION",
                "next_action": {
                    "ok": [
                        "9de8a0c8-0fa5-4a5b-bdfb-ca2e1a1f841e"
                    ]
                },
                "debug_network": false
            }
        ],
        "min_cli_version": "0.0.1",
        "builder_version": "1.0.1"
    }
}