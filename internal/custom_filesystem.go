package internal

import (
	"context"
	"fmt"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"io"
	"strings"
	"syscall"
)

type Root struct {
	fs.Inode
	Client ClientBase
}
type Directory struct {
	fs.Inode

	name string
}

// これは型アサーションをすることでinterfaceの実装ミスをコンパイル時に防ぐために定義している
// 以下詳細：
// (*Root)(nil)でHelloRoot型のnilポインタを返す
// (fs.NodeGetattrer)((*Root)(nil))で↑で作成したHelloRoot型のnilポインタをfs.NodeGetattrer型に型アサーションしようとしている
// こうすることで、HelloRoot構造体が、fs.NodeGetattrer interfaceを実装していない場合にコンパイルエラーが発生するので、コンパイル時に実装ミスに気づける
var _ = (fs.NodeGetattrer)((*Root)(nil))

var _ = (fs.NodeReaddirer)((*Root)(nil))
var _ = (fs.NodeLookuper)((*Root)(nil))

// TODO: ここで指定したデータとlookupなどの各メソッドで定義した情報どちらが使われるのか調べる
func (r *Root) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0755
	return 0
}

func (r *Root) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	isDirectory, err := r.Client.IsDirectory(ctx, name)
	if err != nil {
		return nil, syscall.ENOENT
	}
	fmt.Println(name, isDirectory)
	if isDirectory {
		// ディレクトリの場合
		chile := r.NewInode(ctx, &Directory{name: name}, fs.StableAttr{Mode: syscall.S_IFDIR})

		// ディレクトリであるフラグとファイルパーミッションをORでビット演算している
		out.Mode = syscall.S_IFDIR | 0755
		return chile, 0
	} else {
		// ファイルの場合
		object, err := r.Client.GetObject(ctx, name)
		if err != nil {
			return nil, syscall.ENOENT
		}

		body, err := io.ReadAll(object.Body)
		if err != nil {
			return nil, syscall.ENOENT
		}

		chile := r.NewInode(ctx, &fs.MemRegularFile{
			Data: body,
			Attr: fuse.Attr{
				// TODO: 権限は読み込み書き込み専用でいいか
				Mode: 0444,
			},
		}, fs.StableAttr{Mode: syscall.S_IFREG})

		out.Mode = 0444
		out.Size = uint64(object.ContentLengthByte)
		return chile, 0
	}
}

func (r *Root) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	entry := make([]fuse.DirEntry, 0)
	// rootなのでkey指定しない
	iter, err := r.Client.List(ctx, "")

	if err != nil {
		return nil, syscall.ENOENT
	}
	for i := range iter {
		// / がある場合はディレクトリとみなす
		// child2/grandchild1.txt
		s := strings.Split(iter[i], "/")
		key := s[0]
		if len(s) == 1 {
			// ファイルの場合
			entry = append(entry, fuse.DirEntry{
				Mode: syscall.S_IFREG,
				Name: key,
			})
		} else {
			// ディレクトリの場合
			entry = append(entry, fuse.DirEntry{
				Mode: syscall.S_IFDIR,
				Name: key,
			})
		}
	}

	return fs.NewListDirStream(entry), 0
}

func (d *Directory) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if name == "children.txt" {
		chile := d.NewInode(ctx, &fs.MemRegularFile{
			Data: []byte("childrennnnn"),
			Attr: fuse.Attr{
				Mode: 0444,
			},
		}, fs.StableAttr{Mode: syscall.S_IFREG})

		out.Mode = 0444
		out.Size = 1
		return chile, 0
	}

	return nil, syscall.ENOENT
}

func (d *Directory) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	// TODO: Directoryが保持しているnameでlocalstackを検索してあげる
	//  以下は例として明示的にディレクトリ名とファイル名を指定している

	if d.name == "hoge" {
		entry := []fuse.DirEntry{
			{
				Mode: syscall.S_IFREG,
				Name: "children.txt",
			},
		}
		return fs.NewListDirStream(entry), 0
	}
	return nil, syscall.ENOENT
}