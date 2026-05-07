// Application scenarios:
// - Provide recursive directory upload support for deploy workflows.
// - Mirror local directory structures onto remote Unix-like paths through SFTP.
// - Offer one simple helper for small and medium deployment artifact uploads.
//
// 适用场景：
// - 为部署流程提供递归目录上传能力。
// - 通过 SFTP 把本地目录结构镜像到远端 Unix 风格路径。
// - 为中小规模部署产物上传提供一个简单 helper。
package deploy

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// UploadDir recursively uploads one local directory to the remote directory.
//
// UploadDir 递归把本地目录上传到远端目录。
//
// 中文说明：
// - 它会遍历 localDir 下全部子目录和文件，并在远端按相对路径重建目录树。
// - 远端路径统一按 Unix 风格处理，避免受到本地 Windows 分隔符影响。
// - 当前实现适合 deploy 场景下的中小规模目录覆盖上传。
func UploadDir(client *ssh.Client, localDir, remoteDir string) error {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	localDir = filepath.Clean(localDir)

	return filepath.Walk(localDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(localDir, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			rel = ""
		}

		remotePath := remoteDir
		if rel != "" {
			remotePath = path.Join(remoteDir, rel)
		}
		remotePath = strings.ReplaceAll(remotePath, "\\", "/")

		if info.IsDir() {
			return sftpClient.MkdirAll(remotePath)
		}

		if err := sftpClient.MkdirAll(path.Dir(remotePath)); err != nil {
			return err
		}

		src, err := os.Open(p)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := sftpClient.Create(remotePath)
		if err != nil {
			return fmt.Errorf("create remote %s: %w", remotePath, err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return err
		}
		return nil
	})
}
