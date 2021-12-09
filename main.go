package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	config "github.com/ipfs/go-ipfs-config"
	"github.com/ipfs/go-ipfs/plugin/loader"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
)

func setupPlugins(externalPluginsPath string) error {
	// Load any external plugins if available on externalPluginsPath
	plugins, err := loader.NewPluginLoader(filepath.Join(externalPluginsPath, "plugins"))
	if err != nil {
		return fmt.Errorf("error loading plugins: %s", err)
	}

	// Load preloaded and external plugins
	if err := plugins.Initialize(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := plugins.Inject(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	return nil
}

var rootKey = datastore.NewKey("/local/filesroot")

func main() {
	ctx := context.Background()

	defaultPath, err := config.PathRoot()
	if err != nil {
		log.Fatal(err)
	}

	if err := setupPlugins(defaultPath); err != nil {
		log.Fatal(err)
	}

	repo, err := fsrepo.Open(defaultPath)
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	ds := repo.Datastore()

	root, err := ds.Get(ctx, rootKey)
	if err == datastore.ErrNotFound {
		fmt.Println("empty MFS root")
	} else if err != nil {
		log.Fatal(err)
	} else {
		c, err := cid.Cast(root)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("MFS Root: %s\n", c)
	}

	if len(os.Args) < 2 {
		log.Println("No new MFS root provided")
		return
	}

	fmt.Println("Updating MFS root")
	newRoot, err := cid.Parse(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	err = ds.Put(ctx, rootKey, newRoot.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("MFS root updated to %s\n", newRoot)
}
