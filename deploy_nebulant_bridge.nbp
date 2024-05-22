{
  "blueprint": {
    "cm": {
      "undo": [],
      "redo": []
    },
    "diagram": {
      "cells": [
        {
          "position": {
            "x": 2200,
            "y": 1240
          },
          "type": "nebulant.rectangle.vertical.executionControl.Start",
          "data": {
            "id": "start",
            "version": "1.0.10",
            "provider": "executionControl",
            "settings": {
              "parameters": {
                "input_parameters": [],
                "image": "",
                "description": "",
                "color": "#7986cb",
                "text_color": "#000000",
                "version": "draft",
                "name": "deploy_nebulant_bridge",
                "group_settings_enabled": false
              },
              "info": "This is the start node!",
              "outputs": {}
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "93626e1e-70b3-4b56-bbd5-94b77975256e"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "3a2e9eb3-80f3-4151-9a45-4aefa80b2721"
              }
            ]
          },
          "id": "48992d67-9bda-410d-a3fe-b6391050bc1a",
          "z": 1
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "3fd0dc83-80c3-4de5-9b88-7dd295b68b58",
            "magnet": "circle",
            "port": "a18114a0-a085-4773-9a6d-b0367ba8d5cd"
          },
          "target": {
            "id": "1e8a1f6c-b63f-4955-8b1e-af0fa21f3881",
            "magnet": "circle",
            "port": "4fb36e5a-98cc-4556-aaf1-c8465ff32bae"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "2fe0d4d9-10c5-4f49-abce-38c77527ecf5",
          "z": 2
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "1e8a1f6c-b63f-4955-8b1e-af0fa21f3881",
            "magnet": "circle",
            "port": "9648241a-4cf6-4b5a-9652-23e0682aa775"
          },
          "target": {
            "id": "e4f5a441-6329-4fbf-a128-ed35215726db",
            "magnet": "circle",
            "port": "8b46c13f-8835-4d21-8afe-c9faabcbf125"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "afd83046-1634-4960-9b3d-20a6dca380db",
          "z": 3
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "3eaa02dc-49ec-4977-be75-5298595c1ee1",
            "magnet": "circle",
            "port": "b2b600d9-3772-4553-9d6c-8c87aad84f0f"
          },
          "target": {
            "id": "45023830-eaa9-470c-a71e-efa995bf4f3b",
            "magnet": "circle",
            "port": "3c812f64-920b-4ef9-99c5-458f7eb11d67"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "165336d4-9008-4f81-92ba-1fd30abd2f89",
          "z": 4
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "b8018309-3e2d-4f45-acad-650ada992b7c",
            "magnet": "circle",
            "port": "108dd2df-24e7-47a1-856d-de41ee104c79"
          },
          "target": {
            "id": "58ae9463-5dd7-44d8-b272-5a3e764580df",
            "magnet": "circle",
            "port": "5610b7a1-e6c8-4db8-b1c4-d2078213d9fc"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "7d1023e6-c050-46b8-b180-f5a4c97e6342",
          "z": 5
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "58ae9463-5dd7-44d8-b272-5a3e764580df",
            "magnet": "circle",
            "port": "27a409df-2e0c-4c45-89d3-381fc8d42581"
          },
          "target": {
            "id": "3fd0dc83-80c3-4de5-9b88-7dd295b68b58",
            "magnet": "circle",
            "port": "ecf7b18f-05a6-4cd1-b76f-810853aa2c3c"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "4d09a7a3-032c-4c7c-8183-9117401f0ebf",
          "z": 7
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "24ee50c1-91a2-463f-b5aa-cddd41af5315",
            "magnet": "circle",
            "port": "58c8418a-00c8-4a31-bc8d-ad65f89cc7df"
          },
          "target": {
            "id": "158528a8-d2d1-443d-b0ac-b29d3bf4ae16",
            "magnet": "circle",
            "port": "9a9e5394-1f34-4022-8028-0eedd3f2dbb1"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "06a8eae9-4a9c-4133-8674-783ae25f7036",
          "z": 8
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "0705c312-d9b2-4530-bae2-b6d12d1d21dc",
            "magnet": "circle",
            "port": "528aa4da-ed33-4514-b8c8-fe75e3b1979f"
          },
          "target": {
            "id": "158528a8-d2d1-443d-b0ac-b29d3bf4ae16",
            "magnet": "circle",
            "port": "9a9e5394-1f34-4022-8028-0eedd3f2dbb1"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "b528d989-6581-46ec-aac1-9b9ea05fa585",
          "z": 9
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "158528a8-d2d1-443d-b0ac-b29d3bf4ae16",
            "magnet": "circle",
            "port": "24facd8b-1e64-466e-a1df-599e088a12ed"
          },
          "target": {
            "id": "2de5320e-d3c9-40e7-9a96-80fb9c0b2390",
            "magnet": "circle",
            "port": "d75e88e2-7a3e-4cd7-9b7a-79859e0a84fd"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "77f92459-986c-41af-8366-c424d79e6fd7",
          "z": 10
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "0705c312-d9b2-4530-bae2-b6d12d1d21dc",
            "magnet": "circle",
            "port": "1bd983de-0b41-4cc7-8172-a20e734f4d27"
          },
          "target": {
            "id": "a0c5621f-6a60-4ec9-bdd9-0f0984f8b3ad",
            "magnet": "circle",
            "port": "c4146abc-422d-4af5-83ec-4af600bc8c54"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "bc027c0b-73f3-46dc-b37f-9c8eead78846",
          "z": 11
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "24ee50c1-91a2-463f-b5aa-cddd41af5315",
            "magnet": "circle",
            "port": "730513be-18d6-4879-a9fa-1ace1e35574a"
          },
          "target": {
            "id": "a0c5621f-6a60-4ec9-bdd9-0f0984f8b3ad",
            "magnet": "circle",
            "port": "c4146abc-422d-4af5-83ec-4af600bc8c54"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "0da80999-75b6-46e1-b120-3d6ddba2ef12",
          "z": 12
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "bf84b9ce-d1fe-4995-8aca-a2dd33bb13e6",
            "magnet": "circle",
            "port": "fa2d2539-32b3-449d-9285-7d481885111b"
          },
          "target": {
            "id": "b8018309-3e2d-4f45-acad-650ada992b7c",
            "magnet": "circle",
            "port": "ca1519b0-cc81-429f-8f92-ad4abc8f679c"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "096b4997-027a-42bf-81e2-9476d4a3c753",
          "z": 13
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "9eaf5d61-ea7b-40b0-b270-195f742c671b",
            "magnet": "circle",
            "port": "52f0d84f-67bc-4c51-806c-6729c931f4d7"
          },
          "target": {
            "id": "bf84b9ce-d1fe-4995-8aca-a2dd33bb13e6",
            "magnet": "circle",
            "port": "a4145374-9a02-4655-a3d4-dc406d69b2ef"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "1a67b3c3-615b-45b3-9230-ff1e82ab5124",
          "z": 15
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "c87348df-2b8f-4670-a46c-fc047250fc6e",
            "magnet": "circle",
            "port": "e49a13d2-f849-414c-afee-2e7676a6a9da"
          },
          "target": {
            "id": "bf84b9ce-d1fe-4995-8aca-a2dd33bb13e6",
            "magnet": "circle",
            "port": "a4145374-9a02-4655-a3d4-dc406d69b2ef"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "27a06dd2-6b36-4f1c-88a5-c0d1e619f053",
          "z": 16
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "e4f5a441-6329-4fbf-a128-ed35215726db",
            "magnet": "circle",
            "port": "5b7c9915-136d-4a74-8eb0-cf2b689264d4"
          },
          "target": {
            "id": "7f1d2e32-b555-43f0-83fb-3ad3c63b87f7",
            "magnet": "circle",
            "port": "cbc527d8-9251-460b-a772-e22d72f9c840"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "50ed4cef-06a4-42c2-9de7-71615042cfbc",
          "z": 17
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "3fd0dc83-80c3-4de5-9b88-7dd295b68b58",
            "magnet": "circle",
            "port": "5affc37b-0e12-4c53-a110-0410a4e9544e"
          },
          "target": {
            "id": "b375c2c1-69f2-4faf-a443-b33e31accea3",
            "magnet": "circle",
            "port": "4fb36e5a-98cc-4556-aaf1-c8465ff32bae"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "d994d662-1746-4fe3-8eca-9f6cb6244a2d",
          "z": 18
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "bf84b9ce-d1fe-4995-8aca-a2dd33bb13e6",
            "magnet": "circle",
            "port": "fa2d2539-32b3-449d-9285-7d481885111b"
          },
          "target": {
            "id": "8d9a629e-1386-44fe-8304-d8f7e4858803",
            "magnet": "circle",
            "port": "f9bd3962-2d2a-4fe1-9b95-2318cd00c197"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "f393bc02-b9d1-4bb2-899e-677d4c87eb78",
          "z": 19
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "8d9a629e-1386-44fe-8304-d8f7e4858803",
            "magnet": "circle",
            "port": "cd6b2b77-d2a1-42c3-b4e6-1b59818504d8"
          },
          "target": {
            "id": "58ae9463-5dd7-44d8-b272-5a3e764580df",
            "magnet": "circle",
            "port": "5610b7a1-e6c8-4db8-b1c4-d2078213d9fc"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "ebb7f81c-b213-48b2-a317-dedce8f34be8",
          "z": 20
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "bf84b9ce-d1fe-4995-8aca-a2dd33bb13e6",
            "magnet": "circle",
            "port": "fa2d2539-32b3-449d-9285-7d481885111b"
          },
          "target": {
            "id": "f42adb9a-a7c7-496c-aa33-0bcd8ba36f40",
            "magnet": "circle",
            "port": "8d3da84f-a38b-4fa5-98f0-95996049a0ff"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "43429a09-fc70-4d67-9f50-55ff85eb92b4",
          "z": 21
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "f42adb9a-a7c7-496c-aa33-0bcd8ba36f40",
            "magnet": "circle",
            "port": "a7894bf0-89a3-4796-b9fa-6b75c6eaa19f"
          },
          "target": {
            "id": "58ae9463-5dd7-44d8-b272-5a3e764580df",
            "magnet": "circle",
            "port": "5610b7a1-e6c8-4db8-b1c4-d2078213d9fc"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "31489ced-caee-445e-9d6b-108b2f3a390e",
          "z": 22
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "b375c2c1-69f2-4faf-a443-b33e31accea3",
            "magnet": "circle",
            "port": "9648241a-4cf6-4b5a-9652-23e0682aa775"
          },
          "target": {
            "id": "e4f5a441-6329-4fbf-a128-ed35215726db",
            "magnet": "circle",
            "port": "8b46c13f-8835-4d21-8afe-c9faabcbf125"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "0062c9af-0ab5-4ff5-b1f3-a5503e5386fc",
          "z": 23
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "7f1d2e32-b555-43f0-83fb-3ad3c63b87f7",
            "magnet": "circle",
            "port": "b2b32766-c14b-49c4-b2ab-f5a1852c1ca0"
          },
          "target": {
            "id": "0705c312-d9b2-4530-bae2-b6d12d1d21dc",
            "magnet": "circle",
            "port": "485a46af-dc95-43f7-8fa9-4a3fe225672f"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "ead392ab-f1c0-4990-9b94-147e7a289a19",
          "z": 28
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "a0c5621f-6a60-4ec9-bdd9-0f0984f8b3ad",
            "magnet": "circle",
            "port": "02f57c92-9827-4fe8-aee8-f720f44f0a3f"
          },
          "target": {
            "id": "7183510f-cb2d-4636-9d7d-c7870b7f5433",
            "magnet": "circle",
            "port": "0d90f030-fab0-4413-aa41-3346ff704faf"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "4852d6fa-6f96-469a-b971-1480d83ab189",
          "z": 29
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "7f1d2e32-b555-43f0-83fb-3ad3c63b87f7",
            "magnet": "circle",
            "port": "b2b32766-c14b-49c4-b2ab-f5a1852c1ca0"
          },
          "target": {
            "id": "24ee50c1-91a2-463f-b5aa-cddd41af5315",
            "magnet": "circle",
            "port": "68ca6a02-3b68-4e2e-93ad-e25beaa6c5fb"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "c40ea81e-25cb-4c6c-a0f7-e5e1bc543079",
          "z": 32
        },
        {
          "position": {
            "x": 2201,
            "y": 1664
          },
          "type": "nebulant.rectangle.vertical.executionControl.JoinThreads",
          "data": {
            "id": "join-threads",
            "version": "1.0.0",
            "provider": "executionControl",
            "settings": {
              "parameters": {},
              "info": ""
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "a4145374-9a02-4655-a3d4-dc406d69b2ef"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "fa2d2539-32b3-449d-9285-7d481885111b"
              }
            ]
          },
          "id": "bf84b9ce-d1fe-4995-8aca-a2dd33bb13e6",
          "z": 101
        },
        {
          "position": {
            "x": 2201,
            "y": 2060
          },
          "type": "nebulant.rectangle.vertical.executionControl.JoinThreads",
          "data": {
            "id": "join-threads",
            "version": "1.0.0",
            "provider": "executionControl",
            "settings": {
              "parameters": {},
              "info": ""
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "5610b7a1-e6c8-4db8-b1c4-d2078213d9fc"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "27a409df-2e0c-4c45-89d3-381fc8d42581"
              }
            ]
          },
          "id": "58ae9463-5dd7-44d8-b272-5a3e764580df",
          "z": 105
        },
        {
          "position": {
            "x": 2275,
            "y": 1461
          },
          "type": "nebulant.rectangle.vertical.generic.ReadFile",
          "data": {
            "id": "read-file",
            "version": "1.0.0",
            "provider": "generic",
            "settings": {
              "outputs": {
                "result": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [],
                  "type": "generic:user_variable",
                  "value": "sshkey"
                }
              },
              "parameters": {
                "interpolate": false,
                "file_path": "./nebulant.pem"
              },
              "info": "Leemos la clave de SSH"
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "918a7a86-cd5c-4296-aca4-387222e8f525"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "9f4c8956-c71e-4ccc-a7d6-f0766bc1d834"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "e49a13d2-f849-414c-afee-2e7676a6a9da"
              }
            ]
          },
          "id": "c87348df-2b8f-4670-a46c-fc047250fc6e",
          "z": 119
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "48992d67-9bda-410d-a3fe-b6391050bc1a",
            "magnet": "circle",
            "port": "3a2e9eb3-80f3-4151-9a45-4aefa80b2721"
          },
          "target": {
            "id": "9eaf5d61-ea7b-40b0-b270-195f742c671b",
            "magnet": "circle",
            "port": "656d482f-e9a6-4206-9651-7e1fd2045588"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "c8e466c2-3e93-4fb4-aed6-b8302b8f2e82",
          "z": 133
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "48992d67-9bda-410d-a3fe-b6391050bc1a",
            "magnet": "circle",
            "port": "3a2e9eb3-80f3-4151-9a45-4aefa80b2721"
          },
          "target": {
            "id": "c87348df-2b8f-4670-a46c-fc047250fc6e",
            "magnet": "circle",
            "port": "918a7a86-cd5c-4296-aca4-387222e8f525"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "2f8c58e1-11a4-4919-97ba-5dda428c4dc0",
          "z": 134
        },
        {
          "position": {
            "x": 2221,
            "y": 3334
          },
          "type": "nebulant.rectangle.vertical.executionControl.JoinThreads",
          "data": {
            "id": "join-threads",
            "version": "1.0.0",
            "provider": "executionControl",
            "settings": {
              "parameters": {},
              "info": ""
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "9a9e5394-1f34-4022-8028-0eedd3f2dbb1"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "24facd8b-1e64-466e-a1df-599e088a12ed"
              }
            ]
          },
          "id": "158528a8-d2d1-443d-b0ac-b29d3bf4ae16",
          "z": 135
        },
        {
          "position": {
            "x": 2386,
            "y": 3339
          },
          "type": "nebulant.rectangle.vertical.executionControl.JoinThreads",
          "data": {
            "id": "join-threads",
            "version": "1.0.0",
            "provider": "executionControl",
            "settings": {
              "parameters": {},
              "info": ""
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "c4146abc-422d-4af5-83ec-4af600bc8c54"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "02f57c92-9827-4fe8-aee8-f720f44f0a3f"
              }
            ]
          },
          "id": "a0c5621f-6a60-4ec9-bdd9-0f0984f8b3ad",
          "z": 136
        },
        {
          "position": {
            "x": 2200,
            "y": 1864
          },
          "type": "nebulant.rectangle.vertical.hetznerCloud.FindNetwork",
          "data": {
            "id": "find-network",
            "version": "1.0.0",
            "provider": "hetznerCloud",
            "settings": {
              "outputs": {
                "result": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [],
                  "type": "hetznerCloud:network",
                  "value": "network"
                }
              },
              "parameters": {
                "PerPage": 10,
                "Page": 1,
                "ids": [],
                "Name": "{{ network-name }}",
                "_activeTab": "filters",
                "Filters": [],
                "_maxRetries": 5
              },
              "info": "Buscamos la lan"
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "f9bd3962-2d2a-4fe1-9b95-2318cd00c197"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "c0e2fb8e-f9d1-4d3b-aef1-0afb88912c49"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "cd6b2b77-d2a1-42c3-b4e6-1b59818504d8"
              }
            ]
          },
          "id": "8d9a629e-1386-44fe-8304-d8f7e4858803",
          "z": 157
        },
        {
          "position": {
            "x": 2336,
            "y": 1864
          },
          "type": "nebulant.rectangle.vertical.hetznerCloud.FindServer",
          "data": {
            "id": "find-server",
            "version": "1.0.0",
            "provider": "hetznerCloud",
            "settings": {
              "outputs": {
                "result": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [
                    "ip"
                  ],
                  "type": "hetznerCloud:server",
                  "value": "bastion"
                }
              },
              "parameters": {
                "PerPage": 10,
                "Page": 1,
                "ids": [],
                "Name": "{{ bastion-name }}",
                "_activeTab": "filters",
                "Filters": [],
                "_maxRetries": 5
              },
              "info": "Buscamos el Bastion"
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "ca1519b0-cc81-429f-8f92-ad4abc8f679c"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "317ef89e-bed2-44f9-b49a-83b19819b7f4"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "108dd2df-24e7-47a1-856d-de41ee104c79"
              }
            ]
          },
          "id": "b8018309-3e2d-4f45-acad-650ada992b7c",
          "z": 158
        },
        {
          "position": {
            "x": 2059,
            "y": 1864
          },
          "type": "nebulant.rectangle.vertical.hetznerCloud.FindImage",
          "data": {
            "id": "find-image",
            "version": "1.0.0",
            "provider": "hetznerCloud",
            "settings": {
              "outputs": {
                "result": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [],
                  "type": "hetznerCloud:image",
                  "value": "debian"
                }
              },
              "parameters": {
                "PerPage": 10,
                "Page": 1,
                "_activeTab": "hetzner_images",
                "HetznerImageID": [
                  "114690389"
                ],
                "ImageID": [],
                "Description": "",
                "Name": "",
                "Filters": [
                  {
                    "__uniq": 1715821124110,
                    "name": "Type",
                    "value": [
                      "snapshot",
                      "backup"
                    ]
                  },
                  {
                    "__uniq": 1715821124111,
                    "name": "Status",
                    "value": [
                      "available"
                    ]
                  }
                ],
                "_maxRetries": 5
              },
              "info": "Buscamos la imagen base (Debian 12)"
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "8d3da84f-a38b-4fa5-98f0-95996049a0ff"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "f5d226e7-b00d-4f88-974a-ac26df49f2ef"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "a7894bf0-89a3-4796-b9fa-6b75c6eaa19f"
              }
            ]
          },
          "id": "f42adb9a-a7c7-496c-aa33-0bcd8ba36f40",
          "z": 159
        },
        {
          "position": {
            "x": 2203,
            "y": 2275
          },
          "type": "nebulant.rectangle.vertical.hetznerCloud.FindServer",
          "data": {
            "id": "find-server",
            "version": "1.0.0",
            "provider": "hetznerCloud",
            "settings": {
              "outputs": {
                "result": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [
                    "ip"
                  ],
                  "type": "hetznerCloud:server",
                  "value": "old_bridge"
                }
              },
              "parameters": {
                "PerPage": 10,
                "Page": 1,
                "ids": [],
                "Name": "",
                "_activeTab": "filters",
                "Filters": [
                  {
                    "__uniq": 1715873911969,
                    "name": "LabelSelector",
                    "value": "bridge=true"
                  }
                ],
                "_maxRetries": 5
              },
              "info": "Buscamos el Backend actual"
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "ecf7b18f-05a6-4cd1-b76f-810853aa2c3c"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "5affc37b-0e12-4c53-a110-0410a4e9544e"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "a18114a0-a085-4773-9a6d-b0367ba8d5cd"
              }
            ]
          },
          "id": "3fd0dc83-80c3-4de5-9b88-7dd295b68b58",
          "z": 160
        },
        {
          "position": {
            "x": 2285,
            "y": 2471
          },
          "type": "nebulant.rectangle.vertical.generic.DefineVariables",
          "data": {
            "id": "define-variables",
            "version": "1.0.2",
            "provider": "generic",
            "settings": {
              "info": "Guardamos el Backend actual para borrarlo mas adelante",
              "outputs": {
                "old-bridge-exists": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [],
                  "type": "generic:user_variable",
                  "value": "old-bridge-exists"
                }
              },
              "parameters": {
                "files": [],
                "vars": [
                  {
                    "__uniq": 1715853647251,
                    "name": "new-variable",
                    "value": {
                      "name": "old-bridge-exists",
                      "type": "text",
                      "value": "yes",
                      "required": false,
                      "ask_at_runtime": false,
                      "stack": false
                    }
                  }
                ]
              }
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "4fb36e5a-98cc-4556-aaf1-c8465ff32bae"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "d9bf7b19-a95e-4b70-b927-1eac70470533"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "9648241a-4cf6-4b5a-9652-23e0682aa775"
              }
            ]
          },
          "id": "1e8a1f6c-b63f-4955-8b1e-af0fa21f3881",
          "z": 162
        },
        {
          "position": {
            "x": 2212,
            "y": 2892
          },
          "type": "nebulant.rectangle.vertical.hetznerCloud.FindServer",
          "data": {
            "id": "find-server",
            "version": "1.0.0",
            "provider": "hetznerCloud",
            "settings": {
              "outputs": {
                "result": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [
                    "ip"
                  ],
                  "type": "hetznerCloud:server",
                  "value": "new_bridge"
                }
              },
              "info": "Refrescamos los datos del Backend",
              "parameters": {
                "PerPage": 10,
                "Page": 1,
                "ids": [
                  "{{new_bridge}}"
                ],
                "Name": "",
                "_activeTab": "id",
                "Filters": [],
                "_maxRetries": 5
              }
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "cbc527d8-9251-460b-a772-e22d72f9c840"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "a386b8ea-c610-4630-bb0f-36739e265517"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "b2b32766-c14b-49c4-b2ab-f5a1852c1ca0"
              }
            ]
          },
          "id": "7f1d2e32-b555-43f0-83fb-3ad3c63b87f7",
          "z": 164
        },
        {
          "position": {
            "x": 2166,
            "y": 3515
          },
          "type": "nebulant.rectangle.vertical.hetznerCloud.DeleteServer",
          "data": {
            "id": "delete-server",
            "version": "1.0.0",
            "provider": "hetznerCloud",
            "settings": {
              "outputs": {
                "result": {
                  "waiters": [
                    "success"
                  ],
                  "async": false,
                  "capabilities": [],
                  "type": "hetznerCloud:action",
                  "value": "HC_ACTION_2"
                }
              },
              "info": "Algo no ha ido bien, borramos el server que acabamos de crear",
              "parameters": {
                "ServerIds": [
                  "{{new_bridge}}"
                ],
                "_maxRetries": 5
              }
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "d75e88e2-7a3e-4cd7-9b7a-79859e0a84fd"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "4ce7d86b-4a4e-479b-b14e-3cacb84172b0"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "fea2fee7-789c-4994-919d-16e4c1d5590f"
              }
            ]
          },
          "id": "2de5320e-d3c9-40e7-9a96-80fb9c0b2390",
          "z": 167
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "7183510f-cb2d-4636-9d7d-c7870b7f5433",
            "magnet": "circle",
            "port": "f9d85c2c-f1d2-4000-85a7-add8089a52b8"
          },
          "target": {
            "id": "536d8ec4-bb78-478d-9d5f-64ac2dc09809",
            "magnet": "circle",
            "port": "36ce4e2b-797e-4ccc-80a1-0be3b2aeaa63"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "f9438890-4460-48ff-b63e-47151279ef6a",
          "z": 186
        },
        {
          "position": {
            "x": 2204,
            "y": 2701
          },
          "type": "nebulant.rectangle.vertical.hetznerCloud.CreateServer",
          "data": {
            "id": "create-server",
            "version": "1.0.0",
            "provider": "hetznerCloud",
            "settings": {
              "outputs": {
                "result": {
                  "async": false,
                  "waiters": [
                    "success"
                  ],
                  "capabilities": [
                    "ip"
                  ],
                  "type": "hetznerCloud:server",
                  "value": "new_bridge"
                }
              },
              "parameters": {
                "Labels": [
                  {
                    "__uniq": 1715874220735,
                    "name": "label",
                    "value": [
                      "bridge",
                      "true"
                    ]
                  }
                ],
                "PublicNet": {
                  "_autoAssignIPv6": true,
                  "EnableIPv6": false,
                  "IPv6": [],
                  "_autoAssignIPv4": true,
                  "EnableIPv4": false,
                  "IPv4": []
                },
                "NetworkIds": [
                  "{{network}}"
                ],
                "UserData": "#cloud-config\n",
                "Locations": [
                  "fsn1"
                ],
                "SshKeys": [
                  "20340981"
                ],
                "ImageIds": [
                  "{{debian}}"
                ],
                "ServerTypes": [
                  "cax11"
                ],
                "Name": "{{ bridge-name }}",
                "_activeTab": "general",
                "_maxRetries": 5
              },
              "info": "Creamos el Backend"
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "8b46c13f-8835-4d21-8afe-c9faabcbf125"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "ee940f85-3d7c-48ed-ab30-05044cdf7a0b"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "5b7c9915-136d-4a74-8eb0-cf2b689264d4"
              }
            ]
          },
          "id": "e4f5a441-6329-4fbf-a128-ed35215726db",
          "z": 191
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "7183510f-cb2d-4636-9d7d-c7870b7f5433",
            "magnet": "circle",
            "port": "0d5e165d-3bd2-4422-9701-01aea0ba4d3d"
          },
          "target": {
            "id": "e1cabfd1-2e44-490d-99c4-5ff5c7e55ad8",
            "magnet": "circle",
            "port": "d75e88e2-7a3e-4cd7-9b7a-79859e0a84fd"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "23c6c75f-4c1b-4170-b834-7bc33aa43aae",
          "z": 198
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "536d8ec4-bb78-478d-9d5f-64ac2dc09809",
            "magnet": "circle",
            "port": "f4d92e22-6f39-49e8-9c92-439a84d13da8"
          },
          "target": {
            "id": "c531a217-7d68-4d04-a99c-b3ff0aeacf34",
            "magnet": "circle",
            "port": "0443d9fb-d974-4537-bf1c-250445b900cf"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "18069c0c-62e4-4292-a403-fecaea71fd3a",
          "z": 202
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "a9e1062d-d774-4fad-8ddf-eca6b44740da",
            "magnet": "circle",
            "port": "df602dc9-21f0-43cf-bfdc-48d17efe6c40"
          },
          "target": {
            "id": "41f353bb-f11e-466a-a024-86dc6d705f21",
            "magnet": "circle",
            "port": "cb9c251e-6ade-4014-b163-af5eb500b80d"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "d3710b5f-02a4-4231-9801-6dee258ff25b",
          "z": 220
        },
        {
          "position": {
            "x": 2213,
            "y": 3113
          },
          "type": "nebulant.rectangle.vertical.generic.UploadFiles",
          "data": {
            "id": "upload-files",
            "version": "1.0.3",
            "provider": "generic",
            "settings": {
              "info": "Subimos el código del proyecto",
              "parameters": {
                "proxies": [
                  {
                    "__uniq": 1715858809965,
                    "name": "new-ssh-config",
                    "value": {
                      "_credentials": "privkey",
                      "target": [
                        "{{ bastion.server.public_net.ipv4.ip }}"
                      ],
                      "username": "admin",
                      "privkeyPath": "",
                      "privkey": "{{ sshkey }}",
                      "passphrase": "",
                      "password": "",
                      "port": 22
                    }
                  }
                ],
                "port": 22,
                "password": "",
                "passphrase": "",
                "privkey": "{{ sshkey }}",
                "privkeyPath": "",
                "username": "root",
                "target": [
                  "{{ new_bridge.server.private_net[0].ip }}"
                ],
                "_credentials": "privkey",
                "paths": [
                  {
                    "__uniq": 1716390693240,
                    "name": "new-path-pair",
                    "value": {
                      "_src_type": "folder",
                      "src": "{{ bridge-path }}",
                      "dest": "/tmp/src",
                      "overwrite": false,
                      "recursive": true
                    }
                  },
                  {
                    "__uniq": 1716390693241,
                    "name": "new-path-pair",
                    "value": {
                      "_src_type": "folder",
                      "src": "{{ config-path }}",
                      "dest": "/tmp/deploy_conf",
                      "overwrite": false,
                      "recursive": true
                    }
                  }
                ],
                "_maxRetries": 5
              }
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "485a46af-dc95-43f7-8fa9-4a3fe225672f"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "528aa4da-ed33-4514-b8c8-fe75e3b1979f"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "1bd983de-0b41-4cc7-8172-a20e734f4d27"
              }
            ]
          },
          "id": "0705c312-d9b2-4530-bae2-b6d12d1d21dc",
          "z": 225
        },
        {
          "position": {
            "x": 2102,
            "y": 2470
          },
          "type": "nebulant.rectangle.vertical.generic.DefineVariables",
          "data": {
            "id": "define-variables",
            "version": "1.0.2",
            "provider": "generic",
            "settings": {
              "info": "El backend no existe. Lo reflejamos en una variable",
              "outputs": {
                "old-bridge-exists": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [],
                  "type": "generic:user_variable",
                  "value": "old-bridge-exists"
                }
              },
              "parameters": {
                "files": [],
                "vars": [
                  {
                    "__uniq": 1715853647251,
                    "name": "new-variable",
                    "value": {
                      "name": "old-bridge-exists",
                      "type": "text",
                      "value": "no",
                      "required": false,
                      "ask_at_runtime": false,
                      "stack": false
                    }
                  }
                ]
              }
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "4fb36e5a-98cc-4556-aaf1-c8465ff32bae"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "d9bf7b19-a95e-4b70-b927-1eac70470533"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "9648241a-4cf6-4b5a-9652-23e0682aa775"
              }
            ]
          },
          "id": "b375c2c1-69f2-4faf-a443-b33e31accea3",
          "z": 226
        },
        {
          "position": {
            "x": 2124,
            "y": 1462
          },
          "type": "nebulant.rectangle.vertical.generic.DefineVariables",
          "data": {
            "id": "define-variables",
            "version": "1.0.2",
            "provider": "generic",
            "settings": {
              "parameters": {
                "files": [],
                "vars": [
                  {
                    "__uniq": 1715811343813,
                    "name": "new-variable",
                    "value": {
                      "name": "bastion-name",
                      "type": "text",
                      "value": "Bastion",
                      "required": false,
                      "ask_at_runtime": false,
                      "stack": false
                    }
                  },
                  {
                    "__uniq": 1715811350049,
                    "name": "new-variable",
                    "value": {
                      "name": "bridge-name",
                      "type": "text",
                      "value": "Bridge-{{ ENV.random }}",
                      "required": false,
                      "ask_at_runtime": false,
                      "stack": false
                    }
                  },
                  {
                    "__uniq": 1715854791225,
                    "name": "new-variable",
                    "value": {
                      "name": "network-name",
                      "type": "text",
                      "value": "nebulant-lan",
                      "required": false,
                      "ask_at_runtime": false,
                      "stack": false
                    }
                  },
                  {
                    "__uniq": 1715858836012,
                    "name": "new-variable",
                    "value": {
                      "name": "bridge-path",
                      "type": "text",
                      "value": "./dist",
                      "required": false,
                      "ask_at_runtime": false,
                      "stack": false
                    }
                  },
                  {
                    "__uniq": 1715967853164,
                    "name": "new-variable",
                    "value": {
                      "name": "netdata-uuid",
                      "type": "text",
                      "value": "eba90f71-d35e-4158-9fd5-ea518c5d97b6",
                      "required": false,
                      "ask_at_runtime": false,
                      "stack": false
                    }
                  },
                  {
                    "__uniq": 1716204642895,
                    "name": "new-variable",
                    "value": {
                      "name": "config-path",
                      "type": "text",
                      "value": "./deploy_conf",
                      "required": false,
                      "ask_at_runtime": false,
                      "stack": false
                    }
                  },
                  {
                    "__uniq": 1716206081492,
                    "name": "new-variable",
                    "value": {
                      "name": "bridge-primary-ip-name",
                      "type": "text",
                      "value": "Bridge",
                      "required": false,
                      "ask_at_runtime": false,
                      "stack": false
                    }
                  }
                ]
              },
              "info": "Definimos variables",
              "outputs": {
                "bridge-primary-ip-name": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [],
                  "type": "generic:user_variable",
                  "value": "bridge-primary-ip-name"
                },
                "config-path": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [],
                  "type": "generic:user_variable",
                  "value": "config-path"
                },
                "netdata-uuid": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [],
                  "type": "generic:user_variable",
                  "value": "netdata-uuid"
                },
                "bridge-path": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [],
                  "type": "generic:user_variable",
                  "value": "bridge-path"
                },
                "network-name": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [],
                  "type": "generic:user_variable",
                  "value": "network-name"
                },
                "bridge-name": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [],
                  "type": "generic:user_variable",
                  "value": "bridge-name"
                },
                "bastion-name": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [],
                  "type": "generic:user_variable",
                  "value": "bastion-name"
                }
              }
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "656d482f-e9a6-4206-9651-7e1fd2045588"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "0656d7ef-aee8-491b-837b-2aab799cce1a"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "52f0d84f-67bc-4c51-806c-6729c931f4d7"
              }
            ]
          },
          "id": "9eaf5d61-ea7b-40b0-b270-195f742c671b",
          "z": 227
        },
        {
          "position": {
            "x": 2362,
            "y": 3115
          },
          "type": "nebulant.rectangle.vertical.generic.RunCommand",
          "data": {
            "id": "run-command",
            "version": "1.0.13",
            "provider": "generic",
            "settings": {
              "outputs": {
                "result": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [],
                  "type": "generic:script_execution",
                  "value": "RUN_COMMAND_RESULT"
                }
              },
              "info": "Provisionamos el nuevo server",
              "parameters": {
                "upload_to_remote_target": true,
                "dump_json": false,
                "vars": [],
                "open_dbg_shell_onerror": false,
                "open_dbg_shell_after": false,
                "open_dbg_shell_before": false,
                "proxies": [
                  {
                    "__uniq": 1715858664287,
                    "name": "new-ssh-config",
                    "value": {
                      "_credentials": "privkey",
                      "target": [
                        "{{ bastion.server.public_net.ipv4.ip }}"
                      ],
                      "username": "admin",
                      "privkeyPath": "",
                      "privkey": "{{ sshkey }}",
                      "passphrase": "",
                      "password": "",
                      "port": 22
                    }
                  }
                ],
                "port": 22,
                "password": "",
                "passphrase": "",
                "privkey": "{{ sshkey }}",
                "privkeyPath": "",
                "username": "root",
                "target": [
                  "{{ new_bridge.server.private_net[0].ip }}"
                ],
                "_credentials": "privkey",
                "_run_on_remote": true,
                "scriptParameters": "",
                "scriptName": "",
                "script": "#!/bin/bash\n\nset -e\nset -u\nset -o pipefail\nset -x\n\n# Add user 'admin' to the sudo group\nuseradd -m -s /bin/bash admin\nusermod -aG sudo admin\n\n# Allow sudo without password for the 'admin' user\necho 'admin ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/admin\n\n# Disable password authentication for SSH\nsed -i 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config\n\n# Route the traffic through the Bastion\nip route add default via 172.16.0.1\necho \"nameserver 1.1.1.1\" >> /etc/resolvconf/resolv.conf.d/head\necho \"nameserver 8.8.8.8\" >> /etc/resolvconf/resolv.conf.d/head\nresolvconf -u\ncat <<'EOF' >> /etc/network/interfaces\n  auto enp7s0\n  iface enp7s0 inet dhcp\n      post-up ip route add default via 172.16.0.1\nEOF\n\n# Update and upgrade packages\napt update && apt upgrade -y\n\n# Set locale\nupdate-locale LANG=en_US.UTF-8\n\n# Set timezone\ntimedatectl set-timezone Etc/UTC\n\n# Create swap file\nfallocate -l 2G /swap\nchmod 600 /swap\nmkswap /swap\nswapon /swap\n\n# Make sure the swap file is mounted on boot\necho '/swap none swap sw 0 0' >> /etc/fstab\n\n# SSH\nmkdir -p /home/admin/.ssh\ncp /root/.ssh/authorized_keys /home/admin/.ssh/\nchown -R admin:admin /home/admin/.ssh\nchmod 700 /home/admin/.ssh\nchmod 600 /home/admin/.ssh/authorized_keys\nsed -i -e '/^\\(#\\|\\)PermitRootLogin/s/^.*$/PermitRootLogin no/' /etc/ssh/sshd_config\nsed -i -e '/^\\(#\\|\\)PasswordAuthentication/s/^.*$/PasswordAuthentication no/' /etc/ssh/sshd_config\nsed -i -e '/^\\(#\\|\\)KbdInteractiveAuthentication/s/^.*$/KbdInteractiveAuthentication no/' /etc/ssh/sshd_config\nsed -i -e '/^\\(#\\|\\)ChallengeResponseAuthentication/s/^.*$/ChallengeResponseAuthentication no/' /etc/ssh/sshd_config\nsed -i -e '/^\\(#\\|\\)MaxAuthTries/s/^.*$/MaxAuthTries 10/' /etc/ssh/sshd_config\nsed -i -e '/^\\(#\\|\\)X11Forwarding/s/^.*$/X11Forwarding no/' /etc/ssh/sshd_config\nsed -i -e '/^\\(#\\|\\)AllowAgentForwarding/s/^.*$/AllowAgentForwarding no/' /etc/ssh/sshd_config\nsed -i -e '/^\\(#\\|\\)AuthorizedKeysFile/s/^.*$/AuthorizedKeysFile .ssh\\/authorized_keys/' /etc/ssh/sshd_config\nsed -i '$a AllowUsers admin' /etc/ssh/sshd_config\nsystemctl restart sshd\n\n# Netdata\ncurl https://get.netdata.cloud/kickstart.sh > /tmp/netdata-kickstart.sh && sh /tmp/netdata-kickstart.sh --stable-channel\ncp /usr/lib/netdata/conf.d/stream.conf /etc/netdata/stream.conf\nsed -i '/^\\[stream\\]/,/^\\[/ s/\\(^\\s*enabled\\s*=\\s*\\)no/\\1yes/' /etc/netdata/stream.conf\nsed -i '/^\\[stream\\]/,/^\\[/ s/\\(^\\s*destination\\s*=\\s*\\)/\\1 {{ bastion.server.private_net[0].ip }}/' /etc/netdata/stream.conf\nsed -i '/^\\[stream\\]/,/^\\[/ s/\\(^\\s*api key\\s*=\\s*\\)/\\1 {{ netdata-uuid }}/' /etc/netdata/stream.conf\nsystemctl restart netdata",
                "command": "",
                "pass_to_entrypoint_as_single_param": false,
                "entrypoint": "",
                "_custom_entrypoint": false,
                "_type": "script",
                "_maxRetries": 5
              }
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "68ca6a02-3b68-4e2e-93ad-e25beaa6c5fb"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "58c8418a-00c8-4a31-bc8d-ad65f89cc7df"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "730513be-18d6-4879-a9fa-1ace1e35574a"
              }
            ]
          },
          "id": "24ee50c1-91a2-463f-b5aa-cddd41af5315",
          "z": 228
        },
        {
          "position": {
            "x": 2444,
            "y": 3551
          },
          "type": "nebulant.rectangle.vertical.generic.RunCommand",
          "data": {
            "id": "run-command",
            "version": "1.0.13",
            "provider": "generic",
            "settings": {
              "outputs": {
                "result": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [],
                  "type": "generic:script_execution",
                  "value": "RUN_COMMAND_RESULT_1"
                }
              },
              "info": "Configuración común tanto para el Backend como para los Workers",
              "parameters": {
                "upload_to_remote_target": true,
                "dump_json": false,
                "vars": [],
                "open_dbg_shell_onerror": false,
                "open_dbg_shell_after": false,
                "open_dbg_shell_before": false,
                "proxies": [
                  {
                    "__uniq": 1715961984332,
                    "name": "new-ssh-config",
                    "value": {
                      "_credentials": "privkey",
                      "target": [
                        "{{ bastion.server.public_net.ipv4.ip }}"
                      ],
                      "username": "admin",
                      "privkeyPath": "",
                      "privkey": "{{ sshkey }}",
                      "passphrase": "",
                      "password": "",
                      "port": 22
                    }
                  }
                ],
                "port": 22,
                "password": "",
                "passphrase": "",
                "privkey": "{{ sshkey }}",
                "privkeyPath": "",
                "username": "admin",
                "target": [
                  "{{ new_bridge.server.private_net[0].ip }}"
                ],
                "_credentials": "privkey",
                "_run_on_remote": true,
                "scriptParameters": "",
                "scriptName": "",
                "script": "#!/bin/bash\n\nset -e\nset -u\nset -o pipefail\nset -x\n\n# apt noninteractive\nexport DEBIAN_FRONTEND=noninteractive\n\n# Install\nsudo apt-get -o DPkg::Lock::Timeout=60 update\nsudo apt-mark hold grub*\nsudo apt-get -y full-upgrade\nsudo apt-get -y install libterm-readline-perl-perl\nsudo apt-get install --no-install-recommends -y rsync unzip htop lsof jq curl wget strace traceroute build-essential git acl nano vim gettext locales-all\n\n# Add noatime - http://archive.is/m9X7x#selection-345.0-455.311\nsudo sed -i -r 's/(ext[2-4]\\s+)rw/\\1rw,noatime/g' /etc/fstab\n\n# Remove discard - we're running on SSD\nsudo sed -i -r 's/discard,//g' /etc/fstab\nsudo systemctl enable fstrim.timer\n\n# Instalamos y configuramos nginx\nsudo apt-get install --no-install-recommends -y nginx\nsudo rm -rf /etc/nginx/nginx.conf\nsudo rsync -r /tmp/deploy_conf/nginx/ /etc/nginx/\n\n# Add the \"admin\" user to the \"www-data\" group\nsudo usermod -a -G www-data admin\n\n# Install the project here\nsudo mkdir -p /var/www/webapp\nsudo cp /tmp/src/nebulant-bridge /var/www/webapp/nebulant-bridge\nsudo cp /tmp/src/.env /var/www/webapp/\n\n# webap perms\nsudo chown root:root /var/www/webapp\n\n# This will change the Default ACL\nsudo setfacl -R -d -m u:admin:rwx /var/www/webapp\nsudo setfacl -R -d -m g:www-data:rx /var/www/webapp\nsudo setfacl -R -d -m o::--- /var/www/webapp\n\n# This will change the current ACL\nsudo setfacl -R -m u:admin:rwx /var/www/webapp\nsudo setfacl -R -m g:www-data:rx /var/www/webapp\nsudo setfacl -R -m o::--- /var/www/webapp\n\n# Start the systemd service\nsudo cp /tmp/deploy_conf/systemd/bridge.service /etc/systemd/system/bridge.service\nsudo systemctl enable bridge\nsudo systemctl start bridge\n\nsudo systemctl restart nginx",
                "command": "",
                "pass_to_entrypoint_as_single_param": false,
                "entrypoint": "",
                "_custom_entrypoint": false,
                "_type": "script",
                "_maxRetries": 5
              }
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "0d90f030-fab0-4413-aa41-3346ff704faf"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "0d5e165d-3bd2-4422-9701-01aea0ba4d3d"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "f9d85c2c-f1d2-4000-85a7-add8089a52b8"
              }
            ]
          },
          "id": "7183510f-cb2d-4636-9d7d-c7870b7f5433",
          "z": 229
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "482bfbaa-9eb1-43a5-b9ba-a581bfa35e30",
            "magnet": "circle",
            "port": "35f57f86-d033-4eae-ae08-0ab2a519467c"
          },
          "target": {
            "id": "eb84303a-19ad-4633-8b95-b23183cafd9e",
            "magnet": "circle",
            "port": "b930ce21-e3f5-41bb-b184-d83d187e4f13"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "88f6d7b7-ade0-43f7-8992-3ed49c8b54d8",
          "z": 246
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "eb84303a-19ad-4633-8b95-b23183cafd9e",
            "magnet": "circle",
            "port": "e08b48d6-95dd-40c2-8b6d-83e4554835f2"
          },
          "target": {
            "id": "a9e1062d-d774-4fad-8ddf-eca6b44740da",
            "magnet": "circle",
            "port": "2399a66b-208c-4249-8cf3-1320169f16dd"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "13feb5f3-5e43-440d-b188-007ad8ddd2aa",
          "z": 252
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "c531a217-7d68-4d04-a99c-b3ff0aeacf34",
            "magnet": "circle",
            "port": "e4985c99-71e1-4bf4-ba2c-80638abf4ba5"
          },
          "target": {
            "id": "482bfbaa-9eb1-43a5-b9ba-a581bfa35e30",
            "magnet": "circle",
            "port": "87a341ba-a753-45de-a6c0-20ad806517f8"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "9efafc51-a454-4640-9cfe-12dc2035f568",
          "z": 254
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "482bfbaa-9eb1-43a5-b9ba-a581bfa35e30",
            "magnet": "circle",
            "port": "33173901-5364-449f-9e82-b500b10e8042"
          },
          "target": {
            "id": "eb84303a-19ad-4633-8b95-b23183cafd9e",
            "magnet": "circle",
            "port": "b930ce21-e3f5-41bb-b184-d83d187e4f13"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "67acf5a9-5139-4760-a5c0-01940b01feaf",
          "z": 255
        },
        {
          "position": {
            "x": 2277,
            "y": 3772
          },
          "type": "nebulant.rectangle.vertical.hetznerCloud.DeleteServer",
          "data": {
            "id": "delete-server",
            "version": "1.0.0",
            "provider": "hetznerCloud",
            "settings": {
              "outputs": {
                "result": {
                  "waiters": [
                    "success"
                  ],
                  "async": false,
                  "capabilities": [],
                  "type": "hetznerCloud:action",
                  "value": "HC_ACTION_2"
                }
              },
              "info": "Algo no ha ido bien, borramos el server que acabamos de crear",
              "parameters": {
                "ServerIds": [
                  "{{new_bridge}}"
                ],
                "_maxRetries": 5
              }
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "d75e88e2-7a3e-4cd7-9b7a-79859e0a84fd"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "4ce7d86b-4a4e-479b-b14e-3cacb84172b0"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "fea2fee7-789c-4994-919d-16e4c1d5590f"
              }
            ]
          },
          "id": "e1cabfd1-2e44-490d-99c4-5ff5c7e55ad8",
          "z": 260
        },
        {
          "type": "nebulant.link.Smart",
          "source": {
            "id": "41f353bb-f11e-466a-a024-86dc6d705f21",
            "magnet": "circle",
            "port": "2c6297d9-a13e-47a8-a26f-e5e5c7d46498"
          },
          "target": {
            "id": "3eaa02dc-49ec-4977-be75-5298595c1ee1",
            "magnet": "circle",
            "port": "3d7de3b3-e33b-43a3-b845-1939348f49aa"
          },
          "router": {
            "name": "manhattan",
            "args": {
              "maximumLoops": 10000,
              "maxAllowedDirectionChange": 180,
              "startDirections": [
                "bottom"
              ],
              "endDirections": [
                "top"
              ],
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
          "id": "4bd75535-2c75-4be9-a2f5-924e9c96d554",
          "z": 264
        },
        {
          "position": {
            "x": 2662,
            "y": 4184
          },
          "type": "nebulant.rectangle.vertical.hetznerCloud.UnassignPrimaryIP",
          "data": {
            "id": "unassign-primary-ip",
            "version": "1.0.0",
            "provider": "hetznerCloud",
            "settings": {
              "outputs": {
                "result": {
                  "waiters": [
                    "success"
                  ],
                  "async": false,
                  "capabilities": [],
                  "type": "hetznerCloud:action",
                  "value": "HC_ACTION_3"
                }
              },
              "parameters": {
                "PrimaryIpIds": [
                  "{{bridge_primary_ip}}"
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
                "id": "87a341ba-a753-45de-a6c0-20ad806517f8"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "33173901-5364-449f-9e82-b500b10e8042"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "35f57f86-d033-4eae-ae08-0ab2a519467c"
              }
            ]
          },
          "id": "482bfbaa-9eb1-43a5-b9ba-a581bfa35e30",
          "z": 269
        },
        {
          "position": {
            "x": 2663,
            "y": 4385
          },
          "type": "nebulant.rectangle.vertical.hetznerCloud.AssignPrimaryIP",
          "data": {
            "id": "assign-primary-ip",
            "version": "1.0.0",
            "provider": "hetznerCloud",
            "settings": {
              "outputs": {
                "result": {
                  "async": false,
                  "waiters": [
                    "success"
                  ],
                  "capabilities": [],
                  "type": "hetznerCloud:action",
                  "value": "HC_ACTION"
                }
              },
              "parameters": {
                "ServerIds": [
                  "{{new_bridge}}"
                ],
                "PrimaryIpIds": [
                  "{{bridge_primary_ip}}"
                ],
                "_maxRetries": 5
              },
              "info": "Assign the Bridge IP"
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "b930ce21-e3f5-41bb-b184-d83d187e4f13"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "8ebaf729-5b07-46cc-a91d-43c3f36c0831"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "e08b48d6-95dd-40c2-8b6d-83e4554835f2"
              }
            ]
          },
          "id": "eb84303a-19ad-4633-8b95-b23183cafd9e",
          "z": 270
        },
        {
          "position": {
            "x": 2678,
            "y": 4752
          },
          "type": "nebulant.rectangle.vertical.hetznerCloud.StartServer",
          "data": {
            "id": "start-server",
            "version": "1.0.0",
            "provider": "hetznerCloud",
            "settings": {
              "outputs": {
                "result": {
                  "async": false,
                  "waiters": [
                    "success"
                  ],
                  "capabilities": [],
                  "type": "hetznerCloud:action",
                  "value": "HC_ACTION"
                }
              },
              "parameters": {
                "ids": [
                  "{{new_bridge}}"
                ],
                "_maxRetries": 5
              },
              "info": "Start the server"
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "cb9c251e-6ade-4014-b163-af5eb500b80d"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "63d75921-b7c5-45e7-8720-7988451d672c"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "2c6297d9-a13e-47a8-a26f-e5e5c7d46498"
              }
            ]
          },
          "id": "41f353bb-f11e-466a-a024-86dc6d705f21",
          "z": 271
        },
        {
          "position": {
            "x": 2675,
            "y": 4573
          },
          "type": "nebulant.rectangle.vertical.executionControl.Sleep",
          "data": {
            "id": "sleep",
            "version": "1.0.0",
            "provider": "executionControl",
            "settings": {
              "parameters": {
                "seconds": 15
              },
              "info": "Hetzner keep the server in \"locked\" state a few seconds after assigning a primary IP to it"
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "2399a66b-208c-4249-8cf3-1320169f16dd"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "df602dc9-21f0-43cf-bfdc-48d17efe6c40"
              }
            ]
          },
          "id": "a9e1062d-d774-4fad-8ddf-eca6b44740da",
          "z": 272
        },
        {
          "position": {
            "x": 2684,
            "y": 4948
          },
          "type": "nebulant.threepstar.vertical.executionControl.Condition",
          "data": {
            "id": "condition",
            "version": "1.0.0",
            "provider": "executionControl",
            "settings": {
              "parameters": {
                "conditions_cli": {
                  "rules": []
                },
                "conditions": {
                  "rules": [
                    {
                      "id": "bf2144e9-e4ff-4018-8ebd-a6af501eeb29",
                      "field": "{{old-bridge-exists}}",
                      "operator": "=",
                      "valueSource": "value",
                      "value": "yes"
                    }
                  ],
                  "id": "f2a92e38-5a3f-48ff-95f5-881511cc1b2a"
                },
                "conditions_nonic": {
                  "rules": [
                    {
                      "id": "bf2144e9-e4ff-4018-8ebd-a6af501eeb29",
                      "field": "{{old-bridge-exists}}",
                      "operator": "=",
                      "valueSource": "value",
                      "value": "yes"
                    }
                  ],
                  "id": "f2a92e38-5a3f-48ff-95f5-881511cc1b2a",
                  "combinator": "and"
                }
              },
              "info": "Si había un viejo server..."
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "3d7de3b3-e33b-43a3-b845-1939348f49aa"
              },
              {
                "group": "out-false",
                "attrs": {},
                "id": "688cc296-30eb-4fb9-a1f0-8d03ef90b715"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "ec71b379-6258-4709-bd2e-5bcc3603f280"
              },
              {
                "group": "out-true",
                "attrs": {},
                "id": "b2b600d9-3772-4553-9d6c-8c87aad84f0f"
              }
            ]
          },
          "id": "3eaa02dc-49ec-4977-be75-5298595c1ee1",
          "z": 273
        },
        {
          "position": {
            "x": 2757,
            "y": 5141
          },
          "type": "nebulant.rectangle.vertical.hetznerCloud.DeleteServer",
          "data": {
            "id": "delete-server",
            "version": "1.0.0",
            "provider": "hetznerCloud",
            "settings": {
              "outputs": {
                "result": {
                  "waiters": [
                    "success"
                  ],
                  "async": false,
                  "capabilities": [],
                  "type": "hetznerCloud:action",
                  "value": "HC_ACTION_1"
                }
              },
              "info": "... lo borramos",
              "parameters": {
                "ServerIds": [
                  "{{old_bridge}}"
                ],
                "_maxRetries": 5
              }
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "3c812f64-920b-4ef9-99c5-458f7eb11d67"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "c8f0ecc9-d336-4fc2-89f1-5fa71493bbfc"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "7459e9cf-56de-4942-aef4-858e39aa2f81"
              }
            ]
          },
          "id": "45023830-eaa9-470c-a71e-efa995bf4f3b",
          "z": 274
        },
        {
          "position": {
            "x": 2640,
            "y": 3971
          },
          "type": "nebulant.rectangle.vertical.hetznerCloud.StopServer",
          "data": {
            "id": "stop-server",
            "version": "1.0.0",
            "provider": "hetznerCloud",
            "settings": {
              "outputs": {
                "result": {
                  "async": false,
                  "waiters": [
                    "success"
                  ],
                  "capabilities": [],
                  "type": "hetznerCloud:action",
                  "value": "HC_ACTION"
                }
              },
              "parameters": {
                "ids": [
                  "{{new_bridge}}"
                ],
                "_maxRetries": 5
              },
              "info": "Stop the server (because it must be stopped before attempting the assign op)"
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "0443d9fb-d974-4537-bf1c-250445b900cf"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "1359ee1e-ff6a-43d0-8548-2a2bfb5efbd7"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "e4985c99-71e1-4bf4-ba2c-80638abf4ba5"
              }
            ]
          },
          "id": "c531a217-7d68-4d04-a99c-b3ff0aeacf34",
          "z": 275
        },
        {
          "position": {
            "x": 2630,
            "y": 3776
          },
          "type": "nebulant.rectangle.vertical.hetznerCloud.FindPrimaryIP",
          "data": {
            "id": "find-primary-ip",
            "version": "1.0.0",
            "provider": "hetznerCloud",
            "settings": {
              "outputs": {
                "result": {
                  "waiters": [],
                  "async": false,
                  "capabilities": [
                    "ip"
                  ],
                  "type": "hetznerCloud:primary_ip",
                  "value": "bridge_primary_ip"
                }
              },
              "parameters": {
                "PerPage": 10,
                "Page": 1,
                "ids": [],
                "Name": "{{ bridge-primary-ip-name }}",
                "_activeTab": "filters",
                "Filters": [],
                "_maxRetries": 5
              },
              "info": "Find the primary IP of the Bridge"
            }
          },
          "ports": {
            "items": [
              {
                "group": "in",
                "attrs": {},
                "id": "36ce4e2b-797e-4ccc-80a1-0be3b2aeaa63"
              },
              {
                "group": "out-ko",
                "attrs": {},
                "id": "486028cd-7c25-4a88-9fde-1f7a0f733fc0"
              },
              {
                "group": "out-ok",
                "attrs": {},
                "id": "f4d92e22-6f39-49e8-9c92-439a84d13da8"
              }
            ]
          },
          "id": "536d8ec4-bb78-478d-9d5f-64ac2dc09809",
          "z": 276
        }
      ],
      "zoom": 0.5808051274471135,
      "x": 2508.586669921875,
      "y": 3789.5670166015625
    },
    "diagram_version": "1.0.7",
    "n_warnings": 3,
    "n_errors": 0,
    "actions": [
      {
        "action_id": "48992d67-9bda-410d-a3fe-b6391050bc1a",
        "provider": "generic",
        "version": "1.0.10",
        "first_action": true,
        "action": "start",
        "next_action": {
          "ok": [
            "9eaf5d61-ea7b-40b0-b270-195f742c671b",
            "c87348df-2b8f-4670-a46c-fc047250fc6e"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "bf84b9ce-d1fe-4995-8aca-a2dd33bb13e6",
        "provider": "generic",
        "version": "1.0.0",
        "action": "join_threads",
        "next_action": {
          "ok": [
            "b8018309-3e2d-4f45-acad-650ada992b7c",
            "8d9a629e-1386-44fe-8304-d8f7e4858803",
            "f42adb9a-a7c7-496c-aa33-0bcd8ba36f40"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "58ae9463-5dd7-44d8-b272-5a3e764580df",
        "provider": "generic",
        "version": "1.0.0",
        "action": "join_threads",
        "next_action": {
          "ok": [
            "3fd0dc83-80c3-4de5-9b88-7dd295b68b58"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "c87348df-2b8f-4670-a46c-fc047250fc6e",
        "provider": "generic",
        "version": "1.0.0",
        "action": "read_file",
        "parameters": {
          "interpolate": false,
          "file_path": "./nebulant.pem"
        },
        "output": "sshkey",
        "next_action": {
          "ok": [
            "bf84b9ce-d1fe-4995-8aca-a2dd33bb13e6"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "158528a8-d2d1-443d-b0ac-b29d3bf4ae16",
        "provider": "generic",
        "version": "1.0.0",
        "action": "join_threads",
        "next_action": {
          "ok": [
            "2de5320e-d3c9-40e7-9a96-80fb9c0b2390"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "a0c5621f-6a60-4ec9-bdd9-0f0984f8b3ad",
        "provider": "generic",
        "version": "1.0.0",
        "action": "join_threads",
        "next_action": {
          "ok": [
            "7183510f-cb2d-4636-9d7d-c7870b7f5433"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "8d9a629e-1386-44fe-8304-d8f7e4858803",
        "provider": "hetznerCloud",
        "version": "1.0.0",
        "action": "findone_network",
        "parameters": {
          "max_retries": 5,
          "Name": "{{ network-name }}"
        },
        "output": "network",
        "next_action": {
          "ok": [
            "58ae9463-5dd7-44d8-b272-5a3e764580df"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "b8018309-3e2d-4f45-acad-650ada992b7c",
        "provider": "hetznerCloud",
        "version": "1.0.0",
        "action": "findone_server",
        "parameters": {
          "max_retries": 5,
          "Name": "{{ bastion-name }}"
        },
        "output": "bastion",
        "next_action": {
          "ok": [
            "58ae9463-5dd7-44d8-b272-5a3e764580df"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "f42adb9a-a7c7-496c-aa33-0bcd8ba36f40",
        "provider": "hetznerCloud",
        "version": "1.0.0",
        "action": "findone_image",
        "parameters": {
          "max_retries": 5,
          "id": "114690389"
        },
        "output": "debian",
        "next_action": {
          "ok": [
            "58ae9463-5dd7-44d8-b272-5a3e764580df"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "3fd0dc83-80c3-4de5-9b88-7dd295b68b58",
        "provider": "hetznerCloud",
        "version": "1.0.0",
        "action": "findone_server",
        "parameters": {
          "max_retries": 5,
          "LabelSelector": "bridge=true"
        },
        "output": "old_bridge",
        "next_action": {
          "ok": [
            "1e8a1f6c-b63f-4955-8b1e-af0fa21f3881"
          ],
          "ko": [
            "b375c2c1-69f2-4faf-a443-b33e31accea3"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "1e8a1f6c-b63f-4955-8b1e-af0fa21f3881",
        "provider": "generic",
        "version": "1.0.2",
        "action": "define_variables",
        "parameters": {
          "vars": [
            {
              "key": "old-bridge-exists",
              "ask_at_runtime": false,
              "required": false,
              "type": "string",
              "stack": false,
              "value": "yes"
            }
          ]
        },
        "next_action": {
          "ok": [
            "e4f5a441-6329-4fbf-a128-ed35215726db"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "7f1d2e32-b555-43f0-83fb-3ad3c63b87f7",
        "provider": "hetznerCloud",
        "version": "1.0.0",
        "action": "findone_server",
        "parameters": {
          "max_retries": 5,
          "id": "{{new_bridge}}"
        },
        "output": "new_bridge",
        "next_action": {
          "ok": [
            "0705c312-d9b2-4530-bae2-b6d12d1d21dc",
            "24ee50c1-91a2-463f-b5aa-cddd41af5315"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "2de5320e-d3c9-40e7-9a96-80fb9c0b2390",
        "provider": "hetznerCloud",
        "version": "1.0.0",
        "action": "delete_server",
        "parameters": {
          "ID": "{{new_bridge}}",
          "_waiters": [
            "success"
          ],
          "max_retries": 5
        },
        "output": "HC_ACTION_2",
        "next_action": {},
        "debug_network": true
      },
      {
        "action_id": "e4f5a441-6329-4fbf-a128-ed35215726db",
        "provider": "hetznerCloud",
        "version": "1.0.0",
        "action": "create_server",
        "parameters": {
          "Name": "{{ bridge-name }}",
          "ServerType": {
            "Name": "cax11"
          },
          "Image": {
            "ID": "{{debian}}"
          },
          "SSHKeys": [
            {
              "ID": "20340981"
            }
          ],
          "Location": {
            "Name": "fsn1"
          },
          "UserData": "#cloud-config\n",
          "PublicNet": {
            "EnableIPv4": false,
            "EnableIPv6": false
          },
          "Networks": [
            {
              "ID": "{{network}}"
            }
          ],
          "Labels": {
            "bridge": "true"
          },
          "_waiters": [
            "success"
          ],
          "max_retries": 5
        },
        "output": "new_bridge",
        "next_action": {
          "ok": [
            "7f1d2e32-b555-43f0-83fb-3ad3c63b87f7"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "0705c312-d9b2-4530-bae2-b6d12d1d21dc",
        "provider": "generic",
        "version": "1.0.3",
        "action": "upload_files",
        "parameters": {
          "paths": [
            {
              "_src_type": "folder",
              "src": "{{ bridge-path }}",
              "dest": "/tmp/src",
              "overwrite": false,
              "recursive": true
            },
            {
              "_src_type": "folder",
              "src": "{{ config-path }}",
              "dest": "/tmp/deploy_conf",
              "overwrite": false,
              "recursive": true
            }
          ],
          "username": "root",
          "port": 22,
          "target": "{{ new_bridge.server.private_net[0].ip }}",
          "privkey": "{{ sshkey }}",
          "proxies": [
            {
              "username": "admin",
              "target": "{{ bastion.server.public_net.ipv4.ip }}",
              "port": 22,
              "privkey": "{{ sshkey }}"
            }
          ],
          "max_retries": 5
        },
        "next_action": {
          "ok": [
            "a0c5621f-6a60-4ec9-bdd9-0f0984f8b3ad"
          ],
          "ko": [
            "158528a8-d2d1-443d-b0ac-b29d3bf4ae16"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "b375c2c1-69f2-4faf-a443-b33e31accea3",
        "provider": "generic",
        "version": "1.0.2",
        "action": "define_variables",
        "parameters": {
          "vars": [
            {
              "key": "old-bridge-exists",
              "ask_at_runtime": false,
              "required": false,
              "type": "string",
              "stack": false,
              "value": "no"
            }
          ]
        },
        "next_action": {
          "ok": [
            "e4f5a441-6329-4fbf-a128-ed35215726db"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "9eaf5d61-ea7b-40b0-b270-195f742c671b",
        "provider": "generic",
        "version": "1.0.2",
        "action": "define_variables",
        "parameters": {
          "vars": [
            {
              "key": "bastion-name",
              "ask_at_runtime": false,
              "required": false,
              "type": "string",
              "stack": false,
              "value": "Bastion"
            },
            {
              "key": "bridge-name",
              "ask_at_runtime": false,
              "required": false,
              "type": "string",
              "stack": false,
              "value": "Bridge-{{ ENV.random }}"
            },
            {
              "key": "network-name",
              "ask_at_runtime": false,
              "required": false,
              "type": "string",
              "stack": false,
              "value": "nebulant-lan"
            },
            {
              "key": "bridge-path",
              "ask_at_runtime": false,
              "required": false,
              "type": "string",
              "stack": false,
              "value": "./dist"
            },
            {
              "key": "netdata-uuid",
              "ask_at_runtime": false,
              "required": false,
              "type": "string",
              "stack": false,
              "value": "eba90f71-d35e-4158-9fd5-ea518c5d97b6"
            },
            {
              "key": "config-path",
              "ask_at_runtime": false,
              "required": false,
              "type": "string",
              "stack": false,
              "value": "./deploy_conf"
            },
            {
              "key": "bridge-primary-ip-name",
              "ask_at_runtime": false,
              "required": false,
              "type": "string",
              "stack": false,
              "value": "Bridge"
            }
          ]
        },
        "next_action": {
          "ok": [
            "bf84b9ce-d1fe-4995-8aca-a2dd33bb13e6"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "24ee50c1-91a2-463f-b5aa-cddd41af5315",
        "provider": "generic",
        "version": "1.0.13",
        "action": "run_script",
        "parameters": {
          "pass_to_entrypoint_as_single_param": false,
          "script": "#!/bin/bash\n\nset -e\nset -u\nset -o pipefail\nset -x\n\n# Add user 'admin' to the sudo group\nuseradd -m -s /bin/bash admin\nusermod -aG sudo admin\n\n# Allow sudo without password for the 'admin' user\necho 'admin ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/admin\n\n# Disable password authentication for SSH\nsed -i 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config\n\n# Route the traffic through the Bastion\nip route add default via 172.16.0.1\necho \"nameserver 1.1.1.1\" >> /etc/resolvconf/resolv.conf.d/head\necho \"nameserver 8.8.8.8\" >> /etc/resolvconf/resolv.conf.d/head\nresolvconf -u\ncat <<'EOF' >> /etc/network/interfaces\n  auto enp7s0\n  iface enp7s0 inet dhcp\n      post-up ip route add default via 172.16.0.1\nEOF\n\n# Update and upgrade packages\napt update && apt upgrade -y\n\n# Set locale\nupdate-locale LANG=en_US.UTF-8\n\n# Set timezone\ntimedatectl set-timezone Etc/UTC\n\n# Create swap file\nfallocate -l 2G /swap\nchmod 600 /swap\nmkswap /swap\nswapon /swap\n\n# Make sure the swap file is mounted on boot\necho '/swap none swap sw 0 0' >> /etc/fstab\n\n# SSH\nmkdir -p /home/admin/.ssh\ncp /root/.ssh/authorized_keys /home/admin/.ssh/\nchown -R admin:admin /home/admin/.ssh\nchmod 700 /home/admin/.ssh\nchmod 600 /home/admin/.ssh/authorized_keys\nsed -i -e '/^\\(#\\|\\)PermitRootLogin/s/^.*$/PermitRootLogin no/' /etc/ssh/sshd_config\nsed -i -e '/^\\(#\\|\\)PasswordAuthentication/s/^.*$/PasswordAuthentication no/' /etc/ssh/sshd_config\nsed -i -e '/^\\(#\\|\\)KbdInteractiveAuthentication/s/^.*$/KbdInteractiveAuthentication no/' /etc/ssh/sshd_config\nsed -i -e '/^\\(#\\|\\)ChallengeResponseAuthentication/s/^.*$/ChallengeResponseAuthentication no/' /etc/ssh/sshd_config\nsed -i -e '/^\\(#\\|\\)MaxAuthTries/s/^.*$/MaxAuthTries 10/' /etc/ssh/sshd_config\nsed -i -e '/^\\(#\\|\\)X11Forwarding/s/^.*$/X11Forwarding no/' /etc/ssh/sshd_config\nsed -i -e '/^\\(#\\|\\)AllowAgentForwarding/s/^.*$/AllowAgentForwarding no/' /etc/ssh/sshd_config\nsed -i -e '/^\\(#\\|\\)AuthorizedKeysFile/s/^.*$/AuthorizedKeysFile .ssh\\/authorized_keys/' /etc/ssh/sshd_config\nsed -i '$a AllowUsers admin' /etc/ssh/sshd_config\nsystemctl restart sshd\n\n# Netdata\ncurl https://get.netdata.cloud/kickstart.sh > /tmp/netdata-kickstart.sh && sh /tmp/netdata-kickstart.sh --stable-channel\ncp /usr/lib/netdata/conf.d/stream.conf /etc/netdata/stream.conf\nsed -i '/^\\[stream\\]/,/^\\[/ s/\\(^\\s*enabled\\s*=\\s*\\)no/\\1yes/' /etc/netdata/stream.conf\nsed -i '/^\\[stream\\]/,/^\\[/ s/\\(^\\s*destination\\s*=\\s*\\)/\\1 {{ bastion.server.private_net[0].ip }}/' /etc/netdata/stream.conf\nsed -i '/^\\[stream\\]/,/^\\[/ s/\\(^\\s*api key\\s*=\\s*\\)/\\1 {{ netdata-uuid }}/' /etc/netdata/stream.conf\nsystemctl restart netdata",
          "target": "{{ new_bridge.server.private_net[0].ip }}",
          "username": "root",
          "port": 22,
          "privkey": "{{ sshkey }}",
          "upload_to_remote_target": true,
          "proxies": [
            {
              "username": "admin",
              "target": "{{ bastion.server.public_net.ipv4.ip }}",
              "port": 22,
              "privkey": "{{ sshkey }}"
            }
          ],
          "open_dbg_shell_before": false,
          "open_dbg_shell_after": false,
          "open_dbg_shell_onerror": false,
          "dump_json": false,
          "max_retries": 5
        },
        "output": "RUN_COMMAND_RESULT",
        "next_action": {
          "ok": [
            "a0c5621f-6a60-4ec9-bdd9-0f0984f8b3ad"
          ],
          "ko": [
            "158528a8-d2d1-443d-b0ac-b29d3bf4ae16"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "7183510f-cb2d-4636-9d7d-c7870b7f5433",
        "provider": "generic",
        "version": "1.0.13",
        "action": "run_script",
        "parameters": {
          "pass_to_entrypoint_as_single_param": false,
          "script": "#!/bin/bash\n\nset -e\nset -u\nset -o pipefail\nset -x\n\n# apt noninteractive\nexport DEBIAN_FRONTEND=noninteractive\n\n# Install\nsudo apt-get -o DPkg::Lock::Timeout=60 update\nsudo apt-mark hold grub*\nsudo apt-get -y full-upgrade\nsudo apt-get -y install libterm-readline-perl-perl\nsudo apt-get install --no-install-recommends -y rsync unzip htop lsof jq curl wget strace traceroute build-essential git acl nano vim gettext locales-all\n\n# Add noatime - http://archive.is/m9X7x#selection-345.0-455.311\nsudo sed -i -r 's/(ext[2-4]\\s+)rw/\\1rw,noatime/g' /etc/fstab\n\n# Remove discard - we're running on SSD\nsudo sed -i -r 's/discard,//g' /etc/fstab\nsudo systemctl enable fstrim.timer\n\n# Instalamos y configuramos nginx\nsudo apt-get install --no-install-recommends -y nginx\nsudo rm -rf /etc/nginx/nginx.conf\nsudo rsync -r /tmp/deploy_conf/nginx/ /etc/nginx/\n\n# Add the \"admin\" user to the \"www-data\" group\nsudo usermod -a -G www-data admin\n\n# Install the project here\nsudo mkdir -p /var/www/webapp\nsudo cp /tmp/src/nebulant-bridge /var/www/webapp/nebulant-bridge\nsudo cp /tmp/src/.env /var/www/webapp/\n\n# webap perms\nsudo chown root:root /var/www/webapp\n\n# This will change the Default ACL\nsudo setfacl -R -d -m u:admin:rwx /var/www/webapp\nsudo setfacl -R -d -m g:www-data:rx /var/www/webapp\nsudo setfacl -R -d -m o::--- /var/www/webapp\n\n# This will change the current ACL\nsudo setfacl -R -m u:admin:rwx /var/www/webapp\nsudo setfacl -R -m g:www-data:rx /var/www/webapp\nsudo setfacl -R -m o::--- /var/www/webapp\n\n# Start the systemd service\nsudo cp /tmp/deploy_conf/systemd/bridge.service /etc/systemd/system/bridge.service\nsudo systemctl enable bridge\nsudo systemctl start bridge\n\nsudo systemctl restart nginx",
          "target": "{{ new_bridge.server.private_net[0].ip }}",
          "username": "admin",
          "port": 22,
          "privkey": "{{ sshkey }}",
          "upload_to_remote_target": true,
          "proxies": [
            {
              "username": "admin",
              "target": "{{ bastion.server.public_net.ipv4.ip }}",
              "port": 22,
              "privkey": "{{ sshkey }}"
            }
          ],
          "open_dbg_shell_before": false,
          "open_dbg_shell_after": false,
          "open_dbg_shell_onerror": false,
          "dump_json": false,
          "max_retries": 5
        },
        "output": "RUN_COMMAND_RESULT_1",
        "next_action": {
          "ok": [
            "536d8ec4-bb78-478d-9d5f-64ac2dc09809"
          ],
          "ko": [
            "e1cabfd1-2e44-490d-99c4-5ff5c7e55ad8"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "e1cabfd1-2e44-490d-99c4-5ff5c7e55ad8",
        "provider": "hetznerCloud",
        "version": "1.0.0",
        "action": "delete_server",
        "parameters": {
          "ID": "{{new_bridge}}",
          "_waiters": [
            "success"
          ],
          "max_retries": 5
        },
        "output": "HC_ACTION_2",
        "next_action": {},
        "debug_network": true
      },
      {
        "action_id": "482bfbaa-9eb1-43a5-b9ba-a581bfa35e30",
        "provider": "hetznerCloud",
        "version": "1.0.0",
        "action": "unassign_primary_ip",
        "parameters": {
          "id": "{{bridge_primary_ip}}",
          "_waiters": [
            "success"
          ],
          "max_retries": 5
        },
        "output": "HC_ACTION_3",
        "next_action": {
          "ok": [
            "eb84303a-19ad-4633-8b95-b23183cafd9e"
          ],
          "ko": [
            "eb84303a-19ad-4633-8b95-b23183cafd9e"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "eb84303a-19ad-4633-8b95-b23183cafd9e",
        "provider": "hetznerCloud",
        "version": "1.0.0",
        "action": "assign_primary_ip",
        "parameters": {
          "ID": "{{bridge_primary_ip}}",
          "AssigneeID": "{{new_bridge}}",
          "_waiters": [
            "success"
          ],
          "max_retries": 5
        },
        "output": "HC_ACTION",
        "next_action": {
          "ok": [
            "a9e1062d-d774-4fad-8ddf-eca6b44740da"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "41f353bb-f11e-466a-a024-86dc6d705f21",
        "provider": "hetznerCloud",
        "version": "1.0.0",
        "action": "start_server",
        "parameters": {
          "ID": "{{new_bridge}}",
          "_waiters": [
            "success"
          ],
          "max_retries": 5
        },
        "output": "HC_ACTION",
        "next_action": {
          "ok": [
            "3eaa02dc-49ec-4977-be75-5298595c1ee1"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "a9e1062d-d774-4fad-8ddf-eca6b44740da",
        "provider": "generic",
        "version": "1.0.0",
        "action": "sleep",
        "parameters": {
          "seconds": 15
        },
        "next_action": {
          "ok": [
            "41f353bb-f11e-466a-a024-86dc6d705f21"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "3eaa02dc-49ec-4977-be75-5298595c1ee1",
        "provider": "generic",
        "version": "1.0.0",
        "action": "condition",
        "parameters": {
          "conditions": {
            "rules": [
              {
                "id": "bf2144e9-e4ff-4018-8ebd-a6af501eeb29",
                "field": "{{old-bridge-exists}}",
                "operator": "=",
                "valueSource": "value",
                "value": "yes"
              }
            ],
            "id": "f2a92e38-5a3f-48ff-95f5-881511cc1b2a",
            "combinator": "and"
          }
        },
        "next_action": {
          "ok": {
            "true": [
              "45023830-eaa9-470c-a71e-efa995bf4f3b"
            ]
          }
        },
        "debug_network": true
      },
      {
        "action_id": "45023830-eaa9-470c-a71e-efa995bf4f3b",
        "provider": "hetznerCloud",
        "version": "1.0.0",
        "action": "delete_server",
        "parameters": {
          "ID": "{{old_bridge}}",
          "_waiters": [
            "success"
          ],
          "max_retries": 5
        },
        "output": "HC_ACTION_1",
        "next_action": {},
        "debug_network": true
      },
      {
        "action_id": "c531a217-7d68-4d04-a99c-b3ff0aeacf34",
        "provider": "hetznerCloud",
        "version": "1.0.0",
        "action": "stop_server",
        "parameters": {
          "ID": "{{new_bridge}}",
          "_waiters": [
            "success"
          ],
          "max_retries": 5
        },
        "output": "HC_ACTION",
        "next_action": {
          "ok": [
            "482bfbaa-9eb1-43a5-b9ba-a581bfa35e30"
          ]
        },
        "debug_network": true
      },
      {
        "action_id": "536d8ec4-bb78-478d-9d5f-64ac2dc09809",
        "provider": "hetznerCloud",
        "version": "1.0.0",
        "action": "findone_primary_ip",
        "parameters": {
          "max_retries": 5,
          "Name": "{{ bridge-primary-ip-name }}"
        },
        "output": "bridge_primary_ip",
        "next_action": {
          "ok": [
            "c531a217-7d68-4d04-a99c-b3ff0aeacf34"
          ]
        },
        "debug_network": true
      }
    ],
    "min_cli_version": "0.0.1",
    "builder_version": "1.1.0"
  }
}