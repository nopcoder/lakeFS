package io.treeverse.clients.examples

import java.net.URI
import java.nio.charset.Charset
import scala.collection.JavaConverters._
import scala.util.{Try,Success,Failure}

import org.apache.hadoop.fs
import org.apache.hadoop.fs.s3a

import software.amazon.awssdk.services.s3.model.ListObjectsRequest
import software.amazon.awssdk.services.s3.model.PutObjectRequest
import software.amazon.awssdk.services.s3.S3Client
import scala.collection.immutable.Stream

/** Exercise an underlying FileSystem. */
class UnderlyingFS(s3aFS: s3a.S3AFileSystem, bucket: String) {
  val utf8 = Charset.forName("UTF-8")

  def upload(key: String, contents: String) = {
    val path = new fs.Path("s3a", bucket, key)
    val os = s3aFS.create(path)
    os.write(contents.getBytes(utf8))
  }

  def download(key: String): String = {
    val path = new fs.Path("s3a", bucket, key)
    val is = s3aFS.open(path, 16384)
    val bytes = Stream.continually({
      val buf = new Array[Byte](10)
      if (is.read(buf) == -1) null else buf
    }).takeWhile(buf => buf != null)
      .fold(new Array[Byte](0))(_ ++ _)
    new String(bytes)
  }
}

/** Validate an object store. */
class ObjectStore(val s3: S3Client, val bucket: String) {
  /** @return keys of all objects that start with prefix */
  def list(prefix: String): Seq[String] = {
    val req = ListObjectsRequest.builder
      .bucket(bucket)
      .prefix(prefix)
      .build
    val res = s3.listObjects(req)
    res.contents.asScala.map(_.key)
  }
}

object S3AUser extends App {
  override def main(args: Array[String]) {
    var numFailures = 0

    if (args.length != 1) {
      Console.err.println("Usage: ... s3://bucket/prefix/to/write")
      System.exit(1)
    }

    Try(new URI(args(0))) match {
      case Failure(e) => {
        Console.err.printf("parse URI: %s", e)
        System.exit(1)
      }
      case Success(baseURI) => {
        if (baseURI.getScheme != "s3") {
          Console.err.printf("got scheme %s but can only handle s3\n", baseURI.getScheme)
          System.exit(1)
        }
        val bucket = baseURI.getHost
        val prefix = baseURI.getPath

        val fs = new UnderlyingFS(new s3a.S3AFileSystem, bucket)

        val s3 = S3Client.builder.build
        val store = new ObjectStore(s3, bucket)

        fs.upload(prefix + "/abc", "quick brown fox")
        fs.upload(prefix + "/d/e/f/xyz", "foo bar")

        val abcContents = fs.download(prefix + "/abc")
        if (abcContents != "quick brown fox") {
          Console.err.printf("got /abc = \"%s\"\n", abcContents)
          numFailures += 1
        }
        val xyzContents = fs.download(prefix + "/d/e/f/xyz")
        if (xyzContents != "foo bar") {
          Console.err.printf("got /abc = \"%s\"\n", abcContents)
          numFailures += 1
        }

        val listing = store.list(prefix).toList
        val expected = scala.List("/abc", "/d/e/f/xyz")
        val tooMany = listing diff expected
        if (!tooMany.isEmpty) {
          Console.err.printf("unexpected objects on S3: %s", tooMany)
          numFailures += 1
        }
        val tooFew = expected diff listing
        if (!tooFew.isEmpty) {
          Console.err.printf("missing objects on S3: %s", tooFew)
          numFailures += 1
        }
      }
    }

    System.exit(if (numFailures > 0) 1 else 0)
  }
}
