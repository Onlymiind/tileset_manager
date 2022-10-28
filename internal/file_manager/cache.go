package file_manager

import (
	"container/list"
	"errors"

	"github.com/Onlymiind/tileset_manager/internal/common"
)

type tileCache struct {
	cache    map[string]common.Tiles
	queue    *list.List
	queueMap map[string]*list.Element
	maxSize  common.MemorySize
	size     common.MemorySize
}

func newTileCache(size common.MemorySize) tileCache {
	return tileCache{
		cache:    map[string]common.Tiles{},
		queue:    list.New(),
		queueMap: map[string]*list.Element{},
		maxSize:  size,
	}
}

func (c *tileCache) getTile(file string, index uint8) ([]byte, error) {
	data, ok := c.cache[file]
	if !ok {
		tiles, err := ExtractTileData(file)
		if err != nil {
			return nil, common.Wrap(err, "cache", "could not get tile data")
		}

		for c.size != 0 && tiles.Size+c.size > c.maxSize {
			front := c.queue.Front()
			name := front.Value.(string)

			if size := c.cache[name].Size; size > c.size {
				c.size = common.MemorySizeFrom(0, common.Bytes)
			} else {
				c.size -= size
			}
			c.queue.Remove(front)
			delete(c.queueMap, name)
			delete(c.cache, name)
		}

		c.queueMap[file] = c.queue.PushBack(file)

		c.cache[file] = *tiles
		data = c.cache[file]
		c.size += tiles.Size
	}

	if int(index) >= len(data.Data) {
		return nil, errors.New("tile index out of bounds")
	}

	c.queue.MoveToBack(c.queueMap[file])

	return data.Data[index], nil
}

func (c tileCache) getSize() common.MemorySize {
	return c.size
}
