# Maintainer: Yishen Miao <mys721tx@gmail.com>
# Packager: Yishen Miao <mys721tx@gmail.com>
pkgname=pgb-git
pkgver=$(make version)
pkgrel=1
pkgdesc="Pythia Gata Bot"
arch=('i686' 'x86_64' 'armv7h' 'armv6h' 'aarch64')
url="https://github.com/mys721tx/pgb"
license=('GPL')
conflicts=('pgb')
provides=("pgb=$pkgver")
makedepends=(
  'git'
  'go'
)

build() {
  make VERSION=$pkgver DESTDIR="$pkgdir" PREFIX=/usr -C "$startdir"
}

package() {
  make VERSION=$pkgver DESTDIR="$pkgdir" PREFIX=/usr -C "$startdir" install
  install -Dm644 ../pgb.service "$pkgdir"/usr/lib/systemd/system/pgb.service
}
