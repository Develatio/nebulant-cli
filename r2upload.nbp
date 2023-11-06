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
                ]
            ],
            "redo": []
        },
        "diagram": {
            "cells": [
                {
                    "position": {
                        "x": 2020,
                        "y": 1020
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
                    "z": 1
                },
                {
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
                },
                {
                    "position": {
                        "x": 2080,
                        "y": 1300
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
                    "z": 4
                }
            ]
        },
        "diagram_version": "1.0.2",
        "n_warnings": 0,
        "n_errors": 0,
        "actions": [
            {
                "action_id": "c5ffc1d6-2f79-4468-826c-76dba2ae4a33",
                "provider": "generic",
                "version": "1.0.10",
                "first_action": true,
                "action": "start",
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
            }
        ],
        "min_cli_version": "0.0.1",
        "builder_version": "1.0.1"
    }
}