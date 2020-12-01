package transform

import (
	"os"
	"fmt"
	"math"
)

//divide un archivo y lo pasa a un arreglo de chunks de bytes. Están ordenados.
func FileToChunks(path string, name string) ([][]byte, error) {
	file, err := os.Open(path+"/"+name);

	if err != nil {
			fmt.Println(err)
			return nil,err;
	}

	defer file.Close()

	fileInfo, _ := file.Stat()
	var fileSize int64 = fileInfo.Size()
	const fileChunk = 250000 // 250 KB

	// calculate total number of parts the file will be chunked into
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	fmt.Printf("Splitting to %d pieces.\n", totalPartsNum)

	chunks:= [][]byte{};

	for i := uint64(0); i < totalPartsNum; i++ {
		partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		partBuffer := make([]byte, partSize)

		file.Read(partBuffer)
		if err != nil {
				fmt.Println(err)
				return nil, err
		}
		chunks=append(chunks, partBuffer);
	}
	fmt.Println("splitted file: "+name)
	return chunks,nil;
}

//Combina arreglo de chunks a un archivo. Se asume que los bytes vienen ordenados
func ChunksToFile(chunks [][]byte, name string, path string) error{
	newFileName := path+"/"+name;
    _, err := os.Create(newFileName)

	if err != nil {
        fmt.Println(err)
        return err
    }

    //set the newFileName file to APPEND MODE!!
    // open files r and w
	file, err := os.OpenFile(newFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    // IMPORTANT! do not defer a file.Close when opening a file for APPEND mode!
	// defer file.Close()
    // just information on which part of the new file we are appending
    //var writePosition int64 = 0

    for j := 0; j < len(chunks); j++ {
		//read a chunk
		n, err := file.Write(chunks[j])
		fmt.Println("chunk ", j ," size: ",len(chunks[j]))

		if err != nil {
				fmt.Println(err)
				os.Exit(1)
		}

		file.Sync() //flush to disk

		// free up the buffer for next cycle
		// should not be a problem if the chunk size is small, but
		// can be resource hogging if the chunk size is huge.
		// also a good practice to clean up your own plate after eating

		//chunkBufferBytes = nil // reset or empty our buffer


		fmt.Println("Recombining part [", j, "] into : ", newFileName)
		fmt.Println("Written ", n, " bytes")

		
    }
     // now, we close the newFileName
	file.Close()
	return nil
}