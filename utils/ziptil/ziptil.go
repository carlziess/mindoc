package ziptil

import (
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/lifei6671/mindoc/utils/filetil"
)

//解压zip文件
//@param			zipFile			需要解压的zip文件
//@param			dest			需要解压到的目录
//@return			err				返回错误
func Unzip(zipFile, dest string) (err error) {
	dest = strings.TrimSuffix(dest, "/") + "/"
	// 打开一个zip格式文件
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer r.Close()
	// 迭代压缩文件中的文件，打印出文件中的内容
	for _, f := range r.File {
		if !f.FileInfo().IsDir() { //非目录，且不包含__MACOSX
			if folder := dest + filepath.Dir(f.Name); !strings.Contains(folder, "__MACOSX") {
				os.MkdirAll(folder, 0777)
				if fcreate, err := os.Create(dest + strings.TrimPrefix(f.Name, "./")); err == nil {
					if rc, err := f.Open(); err == nil {
						io.Copy(fcreate, rc)
						rc.Close() //不要用defer来关闭，如果文件太多的话，会报too many open files 的错误
						fcreate.Close()
					} else {
						fcreate.Close()
						return err
					}
				} else {
					return err
				}
			}
		}
	}
	return nil
}

//压缩指定文件或文件夹
//@param			dest			压缩后的zip文件目标，如/usr/local/hello.zip
//@param			filepath		需要压缩的文件或者文件夹
//@return			err				错误。如果返回错误，则会删除dest文件
func Zip(dest string, filepath ...string) (err error) {
	if len(filepath) == 0 {
		return errors.New("lack of file")
	}
	//创建文件
	fzip, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer fzip.Close()

	var filelist []filetil.FileList
	for _, file := range filepath {
		if info, err := os.Stat(file); err == nil {
			if info.IsDir() { //目录，则扫描文件
				if f, _ := filetil.ScanFiles(file); len(f) > 0 {
					filelist = append(filelist, f...)
				}
			} else { //文件
				filelist = append(filelist, filetil.FileList{
					IsDir: false,
					Name:  info.Name(),
					Path:  file,
				})
			}
		} else {
			return err
		}
	}
	w := zip.NewWriter(fzip)
	defer w.Close()
	for _, file := range filelist {
		if !file.IsDir {
			if fw, err := w.Create(strings.TrimLeft(file.Path, "./")); err != nil {
				return err
			} else {
				if fileContent, err := ioutil.ReadFile(file.Path); err != nil {
					return err
				} else {
					if _, err = fw.Write(fileContent); err != nil {
						return err
					}
				}
			}
		}
	}
	return
}

func Compress(dst string,src string) (err error) {
	d, _ := os.Create(dst)
	defer d.Close()
	w := zip.NewWriter(d)
	defer w.Close()

	src = strings.Replace(src,"\\","/",-1)
	f, err := os.Open(src)

	if err != nil {
		return err
	}

	//prefix := src[strings.LastIndex(src,"/"):]

	err = compress(f, "", w)

	if err != nil {
		return err
	}

	return nil
}


func compress(file *os.File, prefix string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		if prefix != "" {
			prefix = prefix + "/" + info.Name()
		}else{
			prefix = info.Name()
		}
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = compress(f, prefix, zw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := zip.FileInfoHeader(info)
		if prefix != "" {
			header.Name = prefix + "/" + header.Name
		}

		if err != nil {
			return err
		}
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}






