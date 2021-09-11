return {
  version = "1.2",
  luaversion = "5.1",
  tiledversion = "1.2.4",
  orientation = "orthogonal",
  renderorder = "right-down",
  width = 25,
  height = 21,
  tilewidth = 16,
  tileheight = 16,
  nextlayerid = 5,
  nextobjectid = 63,
  properties = {},
  tilesets = {
    {
      name = "forest",
      firstgid = 1,
      tilewidth = 16,
      tileheight = 16,
      spacing = 0,
      margin = 0,
      columns = 25,
      image = "../forest-tiles.png",
      imagewidth = 400,
      imageheight = 400,
      tileoffset = {
        x = 0,
        y = 0
      },
      grid = {
        orientation = "orthogonal",
        width = 16,
        height = 16
      },
      properties = {},
      terrains = {},
      tilecount = 625,
      tiles = {
        {
          id = 80,
          animation = {
            {
              tileid = 80,
              duration = 800
            },
            {
              tileid = 105,
              duration = 800
            }
          }
        }
      }
    }
  },
  layers = {
    {
      type = "tilelayer",
      id = 1,
      name = "background",
      x = 0,
      y = 0,
      width = 25,
      height = 21,
      visible = true,
      opacity = 1,
      offsetx = 0,
      offsety = 0,
      properties = {},
      encoding = "lua",
      data = {
        1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 257, 258, 5, 1, 1, 3, 1, 1, 1, 1, 1, 35, 1,
        1, 1, 1, 1, 1, 4, 1, 1, 1, 1, 1, 1, 257, 258, 1, 81, 1, 2, 1, 1, 1, 1, 1, 35, 1,
        1, 1, 1, 1, 1, 56, 4, 3, 4, 56, 81, 1, 257, 258, 1, 1, 1, 81, 1, 1, 1, 1, 1, 60, 1,
        1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 4, 1, 257, 258, 1, 1, 5, 1, 1, 1, 1, 1, 1, 85, 1,
        1, 1, 1, 1, 1, 1, 1, 1, 206, 206, 5, 1, 257, 258, 56, 1, 5, 4, 1, 1, 110, 111, 1, 110, 1,
        134, 135, 136, 134, 135, 136, 134, 135, 136, 137, 3, 256, 257, 258, 1, 256, 1, 1, 133, 134, 135, 136, 134, 135, 136,
        1, 1, 1, 1, 2, 2, 2, 2, 2, 1, 2, 2, 257, 258, 1, 56, 56, 1, 1, 1, 1, 1, 1, 2, 1,
        1, 56, 3, 5, 81, 3, 1, 5, 3, 56, 31, 2, 257, 258, 1, 2, 81, 5, 2, 2, 2, 2, 2, 2, 1,
        1, 1, 1, 1, 2, 1, 5, 1, 2, 2, 2, 2, 257, 258, 56, 2, 2, 2, 1, 3, 3, 5, 2, 4, 1,
        281, 281, 281, 209, 210, 1, 81, 1, 207, 208, 281, 281, 282, 283, 1, 81, 2, 338, 339, 340, 341, 342, 343, 81, 1,
        306, 306, 306, 234, 235, 1, 1, 1, 232, 233, 306, 306, 307, 308, 1, 1, 2, 31, 31, 81, 1, 31, 31, 2, 56,
        5, 2, 2, 257, 258, 1, 204, 205, 257, 258, 2, 56, 2, 56, 1, 2, 1, 1, 207, 208, 261, 1, 1, 2, 1,
        1, 1, 5, 259, 260, 281, 281, 281, 282, 283, 280, 281, 281, 281, 281, 281, 281, 281, 282, 233, 286, 1, 2, 2, 2,
        1, 81, 1, 284, 285, 285, 306, 306, 307, 308, 305, 306, 306, 306, 306, 306, 306, 306, 307, 308, 31, 2, 2, 2, 2,
        1, 4, 2, 2, 56, 3, 4, 5, 2, 1, 1, 1, 81, 1, 2, 5, 1, 2, 255, 255, 255, 56, 2, 1, 1,
        2, 2, 56, 2, 2, 4, 5, 1, 1, 1, 56, 5, 2, 56, 1, 4, 2, 1, 5, 2, 2, 1, 2, 2147483687, 1,
        1, 1, 56, 1, 2, 2, 5, 81, 2, 2, 2, 2, 2, 2, 1, 2, 2, 2, 4, 56, 56, 1, 2, 1, 108,
        1, 1, 81, 3, 2, 5, 5, 2, 5, 1, 2, 2, 1, 5, 2, 1, 1, 2, 2, 1, 1, 2, 2, 1, 1,
        1, 1, 1, 1, 1, 1, 5, 3, 1, 2, 81, 1, 1, 5, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
        1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 81, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
        1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1
      }
    },
    {
      type = "tilelayer",
      id = 2,
      name = "foreground",
      x = 0,
      y = 0,
      width = 25,
      height = 21,
      visible = true,
      opacity = 1,
      offsetx = 0,
      offsety = 0,
      properties = {},
      encoding = "lua",
      data = {
        34, 35, 36, 34, 35, 36, 34, 35, 36, 62, 63, 26, 51, 51, 26, 26, 26, 57, 58, 59, 35, 36, 34, 57, 58,
        34, 35, 36, 34, 35, 36, 34, 35, 36, 62, 63, 26, 51, 51, 26, 26, 26, 57, 58, 59, 35, 36, 34, 57, 58,
        34, 35, 36, 34, 60, 61, 59, 60, 61, 62, 63, 26, 51, 51, 26, 26, 26, 57, 58, 59, 60, 61, 59, 57, 58,
        84, 85, 86, 84, 85, 86, 84, 85, 86, 87, 88, 26, 51, 51, 26, 26, 26, 82, 83, 84, 85, 86, 84, 57, 58,
        109, 110, 111, 109, 110, 111, 109, 110, 111, 112, 113, 26, 51, 51, 26, 26, 26, 107, 108, 109, 216, 217, 109, 57, 58,
        27, 27, 26, 26, 26, 26, 26, 26, 105, 105, 51, 51, 51, 51, 51, 51, 26, 238, 239, 240, 241, 242, 243, 57, 58,
        27, 27, 27, 27, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 51, 263, 264, 265, 266, 267, 268, 57, 58,
        27, 27, 27, 27, 27, 27, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 288, 289, 290, 291, 292, 293, 57, 58,
        27, 27, 27, 27, 27, 128, 129, 130, 26, 26, 26, 26, 26, 26, 26, 26, 26, 313, 314, 315, 316, 317, 318, 57, 58,
        27, 27, 27, 27, 26, 153, 154, 155, 26, 26, 26, 203, 26, 26, 0, 0, 0, 26, 26, 26, 26, 26, 26, 57, 58,
        27, 27, 27, 27, 27, 178, 179, 180, 26, 26, 26, 26, 26, 26, 0, 0, 0, 0, 0, 0, 0, 0, 0, 57, 58,
        27, 27, 27, 27, 27, 26, 26, 26, 26, 26, 26, 26, 26, 26, 26, 0, 26, 26, 26, 26, 26, 26, 0, 57, 58,
        27, 27, 27, 27, 27, 0, 0, 26, 26, 0, 26, 26, 26, 26, 0, 0, 26, 26, 26, 26, 26, 26, 0, 57, 58,
        27, 27, 27, 27, 27, 26, 26, 0, 0, 0, 0, 0, 0, 26, 26, 26, 26, 26, 26, 26, 26, 26, 0, 57, 58,
        27, 27, 27, 27, 27, 26, 26, 0, 0, 0, 0, 0, 0, 0, 0, 0, 26, 26, 26, 26, 26, 26, 0, 57, 58,
        10, 11, 9, 10, 11, 9, 10, 11, 9, 10, 11, 9, 10, 11, 9, 10, 11, 9, 10, 11, 9, 10, 11, 82, 83,
        35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 10,
        35, 36, 59, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35,
        35, 36, 59, 35, 36, 59, 60, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35,
        35, 36, 59, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35,
        35, 36, 59, 35, 36, 59, 60, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35, 36, 34, 35
      }
    },
    {
      type = "objectgroup",
      id = 3,
      name = "events",
      visible = true,
      opacity = 1,
      offsetx = 0,
      offsety = 0,
      draworder = "topdown",
      properties = {
        ["collidable"] = true
      },
      objects = {
        {
          id = 34,
          name = "goto_forest1",
          type = "touch",
          shape = "rectangle",
          x = -26,
          y = 64,
          width = 26,
          height = 200,
          rotation = 0,
          visible = true,
          properties = {}
        },
        {
          id = 60,
          name = "rival_fight",
          type = "touch",
          shape = "rectangle",
          x = 104,
          y = 91,
          width = 20,
          height = 160,
          rotation = 0,
          visible = true,
          properties = {}
        },
        {
          id = 61,
          name = "goto_forest3",
          type = "touch",
          shape = "rectangle",
          x = 144,
          y = -20,
          width = 150,
          height = 20,
          rotation = 0,
          visible = true,
          properties = {}
        }
      }
    },
    {
      type = "objectgroup",
      id = 4,
      name = "collision",
      visible = true,
      opacity = 1,
      offsetx = 0,
      offsety = 0,
      draworder = "topdown",
      properties = {
        ["collidable"] = true,
        ["solid"] = true
      },
      objects = {
        {
          id = 37,
          name = "",
          type = "",
          shape = "rectangle",
          x = 0,
          y = 0,
          width = 170,
          height = 73,
          rotation = 0,
          visible = true,
          properties = {}
        },
        {
          id = 41,
          name = "",
          type = "",
          shape = "rectangle",
          x = 162,
          y = 114,
          width = 12,
          height = 14,
          rotation = 0,
          visible = true,
          properties = {}
        },
        {
          id = 44,
          name = "",
          type = "",
          shape = "rectangle",
          x = 0,
          y = 251,
          width = 400,
          height = 85,
          rotation = 0,
          visible = true,
          properties = {}
        },
        {
          id = 45,
          name = "",
          type = "",
          shape = "rectangle",
          x = 371,
          y = 0,
          width = 29,
          height = 251,
          rotation = 0,
          visible = true,
          properties = {}
        },
        {
          id = 46,
          name = "",
          type = "",
          shape = "rectangle",
          x = 275,
          y = 0,
          width = 96,
          height = 76.9998,
          rotation = 0,
          visible = true,
          properties = {}
        },
        {
          id = 47,
          name = "",
          type = "",
          shape = "rectangle",
          x = 278,
          y = 84,
          width = 34,
          height = 76,
          rotation = 0,
          visible = true,
          properties = {}
        },
        {
          id = 52,
          name = "",
          type = "",
          shape = "rectangle",
          x = 91,
          y = 149,
          width = 28,
          height = 29,
          rotation = 0,
          visible = true,
          properties = {}
        },
        {
          id = 53,
          name = "",
          type = "",
          shape = "rectangle",
          x = 98,
          y = 178,
          width = 13,
          height = 11,
          rotation = 0,
          visible = true,
          properties = {}
        },
        {
          id = 54,
          name = "",
          type = "",
          shape = "rectangle",
          x = 0,
          y = 73,
          width = 152,
          height = 18,
          rotation = 0,
          visible = true,
          properties = {}
        },
        {
          id = 56,
          name = "",
          type = "",
          shape = "rectangle",
          x = 274,
          y = 162,
          width = 28,
          height = 14,
          rotation = 0,
          visible = true,
          properties = {}
        },
        {
          id = 57,
          name = "",
          type = "",
          shape = "rectangle",
          x = 338,
          y = 162,
          width = 28,
          height = 14,
          rotation = 0,
          visible = true,
          properties = {}
        },
        {
          id = 58,
          name = "",
          type = "",
          shape = "rectangle",
          x = 322,
          y = 210,
          width = 12,
          height = 14,
          rotation = 0,
          visible = true,
          properties = {}
        },
        {
          id = 59,
          name = "",
          type = "",
          shape = "rectangle",
          x = 338,
          y = 84,
          width = 33,
          height = 76,
          rotation = 0,
          visible = true,
          properties = {}
        },
        {
          id = 62,
          name = "",
          type = "",
          shape = "rectangle",
          x = 312,
          y = 136,
          width = 26,
          height = 24,
          rotation = 0,
          visible = true,
          properties = {}
        }
      }
    }
  }
}
