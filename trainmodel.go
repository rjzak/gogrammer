package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/sjwhitworth/golearn/linear_models"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type dataset struct {
	X [][]float64
	Y []float64
}

func (ds *dataset) GetProblem() *linear_models.Problem {
	return linear_models.NewProblem(ds.X, ds.Y, 0)
}

func (ds *dataset) GetAccuracy(model *linear_models.Model) float64 {
	correctCounter := 1
	for index, dataPoint := range ds.X {
		prediction := linear_models.Predict(model, dataPoint)
		if prediction == ds.Y[index] {
			correctCounter++
		}
	}
	return float64(correctCounter) / float64(len(ds.Y))
}

func CreateRegressor(regulariser string, C, epsilon float64) *linear_models.Parameter {
	defaultModelType := linear_models.L2R_LR
	switch regulariser {
	case "l1":
		defaultModelType = linear_models.L1R_LR
	case "l1l2":
		defaultModelType = linear_models.L2R_LR_DUAL
	}
	return linear_models.NewParameter(defaultModelType, C, epsilon)
}

func ConvertDataset(datasetPath string, hasHeader bool) (*dataset, error) {
	X := make([][]float64, 0)
	Y := make([]float64, 0)

	datasetPath = strings.ToLower(datasetPath)
	ext := filepath.Ext(datasetPath)

	dsFile, err := os.Open(datasetPath)
	if err != nil {
		return nil, err
	}
	defer dsFile.Close()

	dsReader := bufio.NewReader(dsFile)

	if ext == ".csv" {
		lineLength := -1
		first := true
		for {
			line, err := dsReader.ReadString('\n')
			if err == io.EOF {
				break
			}
			lineParts := strings.Split(strings.TrimSpace(line), ",")
			if hasHeader && first {
				first = false
				continue
			}
			if lineLength == -1 {
				lineLength = len(lineParts)
			} else {
				if lineLength != len(lineParts) {
					return nil, errors.New(fmt.Sprintf("Found line length mismatch, %d vs. %d", lineLength, len(lineParts)))
				}
			}

			thisLine := make([]float64, lineLength-2)
			for index, item := range lineParts[:len(lineParts)-2] {
				dataPoint, err := strconv.ParseFloat(item, 64)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("Encountered non-float item: %v\n", item))
				}
				thisLine[index] = dataPoint
			}
			if len(thisLine) == 0 {
				_, _ = fmt.Fprintf(os.Stderr, "Warning, encountered empty line parsing %s as CSV\n", datasetPath)
				continue
			}
			X = append(X, thisLine)

			labelPoint, err := strconv.ParseInt(lineParts[len(lineParts)-1], 10, 32)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Encountered non-int label: %s", lineParts[len(lineParts)-1]))
			}
			Y = append(Y, float64(labelPoint))
		}
	} else {
		if ext == ".libsvm" {
			labelIndexMatch, err := regexp.Compile(`(\d+):`)
			if err != nil {
				return nil, err
			}
			lastInt := int64(-1)
			for {
				thisLine, err := dsReader.ReadString('\n')
				if err == io.EOF {
					break
				}
				labelFinder := labelIndexMatch.FindAllString(thisLine, len(thisLine))
				last := strings.Replace(labelFinder[len(labelFinder)-1], ":", "", 1)
				potentialLastInt, err := strconv.ParseInt(last, 10, 64)
				if err != nil {
					return nil, err
				}
				if potentialLastInt > lastInt {
					lastInt = potentialLastInt
				}

			}
			//fmt.Printf("Found: %d\n", lastInt)

			_, _ = dsFile.Seek(0, 0) // Rewind
			dsReader = bufio.NewReader(dsFile)
			intPairsMatch, err := regexp.Compile(`(\d+):(\d+)`)
			for { // Iterate over the file again now that we know the amount of data points
				thisLine, err := dsReader.ReadString('\n')
				if err == io.EOF {
					break
				}
				pairFinder := intPairsMatch.FindAllString(thisLine, len(thisLine))
				labelThisRow := strings.Split(thisLine, ":")[0]
				labelIntThisRow, err := strconv.ParseFloat(labelThisRow, 64)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("Encountered non-float label: %s", labelThisRow))
				}
				Y = append(Y, labelIntThisRow)
				thisDataRow := make([]float64, lastInt+1)
				//fmt.Printf("Label: %s\n", labelThisRow)
				for _, dataPair := range pairFinder {
					dataPairParts := strings.Split(dataPair, ":")
					//fmt.Printf("\tOrig: %s, Index: %s, Data: %s\n", dataPair, dataPairParts[0], dataPairParts[1])

					dataIndex, err := strconv.ParseInt(dataPairParts[0], 10, 64)
					if err != nil {
						return nil, errors.New(fmt.Sprintf("Encountered non-integer label: %s", dataPairParts[0]))
					}

					dataPoint, err := strconv.ParseFloat(dataPairParts[1], 64)
					if err != nil {
						return nil, errors.New(fmt.Sprintf("Encountered non-float data point: %s", dataPairParts[1]))
					}

					//fmt.Printf("\tOrig: %s, Index: %d, Data: %f, Label: %f\n", dataPair, dataIndex, dataPoint, labelIntThisRow)

					thisDataRow[dataIndex] = dataPoint
				}
				X = append(X, thisDataRow)
			}

		} else {
			return nil, errors.New(fmt.Sprintf("Unknown file extension %s", ext))
		}
	}

	ds := dataset{
		X: X,
		Y: Y,
	}
	return &ds, nil
}

func TrainModel(datasetPath string, modelOutput string, csvHasHeader bool, regulariser string, C float64, epsilon float64) {
	if !exists(datasetPath) {
		_, _ = fmt.Fprintf(os.Stderr, "Dataset path %s does not exist.\n", datasetPath)
		return
	}

	dataset, err := ConvertDataset(datasetPath, csvHasHeader)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to parse dataset: %v\n", err)
		return
	}

	if dataset == nil {
		_, _ = fmt.Fprintf(os.Stderr, "Dataset is nil.\n")
		return
	}

	if len(dataset.X) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Dataset has no data!\n")
		return
	}

	if len(dataset.Y) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Dataset has no targets!\n")
		return
	}

	regressor := CreateRegressor(regulariser, C, epsilon)

	model := linear_models.Train(dataset.GetProblem(), regressor)

	accuracy := dataset.GetAccuracy(model) * 100.0
	fmt.Printf("Training accuracy: %1.2f%%\n", accuracy)

	err = linear_models.Export(model, modelOutput)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to save model to %s: %v.\n", modelOutput, err)
	}
}

func EvaluateModel(datasetPath, modelPath string, csvHasHeader bool) {
	if !exists(datasetPath) {
		_, _ = fmt.Fprintf(os.Stderr, "Dataset path %s does not exist.\n", datasetPath)
		return
	}

	if !exists(modelPath) {
		_, _ = fmt.Fprintf(os.Stderr, "Model path %s does not exist.\n", modelPath)
		return
	}

	dataset, err := ConvertDataset(datasetPath, csvHasHeader)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to parse dataset: %v.\n", err)
		return
	}

	if dataset == nil {
		_, _ = fmt.Fprintf(os.Stderr, "Dataset is nil.\n")
		return
	}

	if len(dataset.X) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Dataset has no data!\n")
		return
	}

	if len(dataset.Y) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Dataset has no targets!\n")
		return
	}

	model := linear_models.Model{}
	err = linear_models.Load(&model, modelPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error loading model: %v.\n", err)
		return
	}

	accuracy := dataset.GetAccuracy(&model) * 100.0
	fmt.Printf("Evaluation accuracy: %1.2f%%\n", accuracy)
}
