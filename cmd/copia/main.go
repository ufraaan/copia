package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ufraaan/copia/internal/copia"
)

func main() {
	var db copia.KV
	if err := db.Open(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()

	printHelp()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "quit" || line == "exit" {
			return
		}
		if line == "" {
			continue
		}
		if err := run(&db, line); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(db *copia.KV, line string) error {
	command, rest, _ := strings.Cut(line, " ")
	switch command {
	case "set":
		key, value, ok := strings.Cut(rest, " ")
		if !ok || key == "" {
			return fmt.Errorf("usage: set <key> <value>")
		}
		_, err := db.Set([]byte(key), []byte(value))
		return err
	case "get":
		if rest == "" || strings.Contains(rest, " ") {
			return fmt.Errorf("usage: get <key>")
		}
		value, found, err := db.Get([]byte(rest))
		if err != nil {
			return err
		}
		if !found {
			fmt.Println("not found")
			return nil
		}
		fmt.Println(string(value))
		return nil
	case "del":
		if rest == "" || strings.Contains(rest, " ") {
			return fmt.Errorf("usage: del <key>")
		}
		deleted, err := db.Del([]byte(rest))
		if err == nil && !deleted {
			fmt.Println("not found")
		}
		return err
	case "rename", "copy":
		source, target, ok := strings.Cut(rest, " ")
		if !ok || source == "" || target == "" || strings.Contains(target, " ") {
			return fmt.Errorf("usage: %s <source> <target>", command)
		}
		var changed bool
		var err error
		if command == "rename" {
			changed, err = db.Rename([]byte(source), []byte(target))
		} else {
			changed, err = db.Copy([]byte(source), []byte(target))
		}
		if err == nil && !changed {
			fmt.Println("not found")
		}
		return err
	case "exists":
		if rest == "" || strings.Contains(rest, " ") {
			return fmt.Errorf("usage: exists <key>")
		}
		found, err := db.Exists([]byte(rest))
		if err == nil {
			fmt.Println(found)
		}
		return err
	case "keys":
		keys, err := db.Keys([]byte(rest))
		if err != nil {
			return err
		}
		for _, key := range keys {
			fmt.Println(string(key))
		}
		return nil
	case "count":
		if rest != "" {
			return fmt.Errorf("usage: count")
		}
		count, err := db.Len()
		if err == nil {
			fmt.Println(count)
		}
		return err
	case "clear":
		if rest != "" {
			return fmt.Errorf("usage: clear")
		}
		return db.Clear()
	case "help":
		printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func printHelp() {
	fmt.Println("copia commands:")
	fmt.Println("  set <key> <value>")
	fmt.Println("  get <key>")
	fmt.Println("  del <key>")
	fmt.Println("  rename <source> <target>")
	fmt.Println("  copy <source> <target>")
	fmt.Println("  exists <key>")
	fmt.Println("  keys [prefix]")
	fmt.Println("  count")
	fmt.Println("  clear")
	fmt.Println("  help")
	fmt.Println("  quit")
}
