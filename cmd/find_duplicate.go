package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/gibsn/serato_tools/files"

	"github.com/dhowden/tag"
	"github.com/spf13/cobra"
	"gopkg.in/fatih/set.v0"
)

var fdCommand = &cobra.Command{
	Use:   "fd",
	Short: "Find duplicate music file in a folder",
	Run:   findDuplicate,
}

func init() {
	fdCommand.Flags().StringVarP(&musicDir, "dir", "d", "", "Root directory for your music (required)")
	fdCommand.MarkFlagRequired("dir")

	rootCmd.AddCommand(fdCommand)
}

func findDuplicate(cmd *cobra.Command, args []string) {
	supportedExtension := set.New(set.ThreadSafe)
	supportedExtension.Add(".mp3",
		".ogg",
		".flac",
		".m4a",
		".mp4")
	f := files.ListFiles(musicDir, supportedExtension)
	s := initDuplicateSet(f)
	printDuplicate(s)
}

func initDuplicateSet(f map[string][]string) set.Interface {
	musicSet := set.New(set.ThreadSafe)
	for _, files := range f {
		for _, f := range files {
			file, _ := os.Open(f)
			tags, err := tag.ReadFrom(file)
			if err != nil {
				log.Printf("Error reading tag for %s\nReason %v\n", f, err)
				continue
			}
			mf := &MusicFile{
				Artist: tags.Artist(),
				Title:  tags.Title(),
			}
			fileFromSet := getFromSet(musicSet, mf)
			if fileFromSet != nil {
				fileFromSet.FilePath = append(fileFromSet.FilePath, f)
			} else {
				mf.FilePath = []string{f}
				musicSet.Add(mf)
			}
		}
	}
	return musicSet
}

func printDuplicate(s set.Interface) {
	total := 0
	s.Each(func(f interface{}) bool {
		mf := f.(*MusicFile)
		if len(mf.FilePath) > 1 {
			total = total + 1
			fmt.Printf("----- Found duplicate for %s - %s ------\n", mf.Artist, mf.Title)
			for _, path := range mf.FilePath {
				fmt.Println(path)
			}
			fmt.Println("")
		}
		return true
	})
	fmt.Println("Number of duplicate songs:", total)
}

type MusicFile struct {
	Artist   string
	Title    string
	FilePath []string
}

func getFromSet(s set.Interface, mf *MusicFile) *MusicFile {
	var foundItem *MusicFile
	//@TODO Maybe replace with Set.Has() function
	s.Each(func(f interface{}) bool {
		if mf.isEqual(f.(*MusicFile)) {
			foundItem = f.(*MusicFile)
			return false
		}
		return true
	})
	return foundItem
}

func (m *MusicFile) isEqual(m2 *MusicFile) bool {
	return m.Title == m2.Title && m.Artist == m2.Artist
}
