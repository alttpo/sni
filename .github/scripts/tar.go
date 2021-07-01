package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	stripPrefix := flag.String("strip", ".", "strip path prefix from filenames embedded in tar")
	outFile := flag.String("o", "archive.tar", "output filename of .tar(.gz)")
	doGzip := flag.Bool("z", false, "compress with gzip")
	flag.Parse()

	*stripPrefix = strings.TrimSpace(*stripPrefix)
	*stripPrefix = strings.TrimSuffix(*stripPrefix, "/")

	finalFile := *outFile
	if *doGzip && filepath.Ext(finalFile) != ".gz" {
		finalFile += ".gz"
	}

	finalFileAbs, absErr := filepath.Abs(finalFile)
	if absErr != nil {
		fmt.Fprintf(os.Stderr, "abs: %s\n", absErr)
		finalFileAbs = finalFile
	}

	var err error
	var f *os.File
	f, err = os.OpenFile(finalFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}()

	var w io.WriteCloser
	w = f
	if *doGzip {
		var gz *gzip.Writer
		gz, err = gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		defer func() {
			err = gz.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		}()

		w = gz
	}

	var tw *tar.Writer
	tw = tar.NewWriter(w)
	defer func() {
		err = tw.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}()

	rootPath := flag.Arg(0)
	err = filepath.WalkDir(
		rootPath,
		func(path string, d fs.DirEntry, inerr error) error {
			if inerr != nil {
				return nil
			}

			fi, err := d.Info()
			if err != nil {
				return err
			}

			if path == *stripPrefix {
				return nil
			}
			outPath := strings.TrimPrefix(path, *stripPrefix + "/")
			if outPath == "" {
				return nil
			}
			if outPath == "." || outPath == ".." {
				return nil
			}

			pathAbs, absErr := filepath.Abs(path)
			if absErr != nil {
				fmt.Fprintf(os.Stderr, "abs: %s\n", absErr)
				pathAbs = path
			}

			// skip output file:
			if pathAbs == finalFileAbs {
				return nil
			}

			fmt.Println(outPath)
			link := ""
			if d.Type()&os.ModeSymlink != 0 {
				link, err = os.Readlink(path)
			}

			var f *os.File
			if !d.IsDir() {
				f, err = os.OpenFile(path, os.O_RDONLY, 0644)
				if err != nil {
					return err
				}
				defer f.Close()
			}

			var hdr *tar.Header
			hdr, err = tar.FileInfoHeader(fi, link)
			if err != nil {
				return err
			}
			// override Name to use the given path so it contains the relative path:
			hdr.Name = outPath

			err = tw.WriteHeader(hdr)
			if err != nil {
				return err
			}
			if !d.IsDir() {
				_, err = io.Copy(tw, f)
				if err != nil {
					return err
				}
			}

			return nil
		})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
