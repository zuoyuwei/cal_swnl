package main

import (
	"errors"
	"fmt"

	// "github.com/cdipaolo/goml/base"
	// "github.com/cdipaolo/goml/linear"
	"sort"

	"gonum.org/v1/gonum/stat"

	//"github.com/grd/statistics"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/NOX73/go-neural"
	"github.com/NOX73/go-neural/learn"
	"github.com/NOX73/go-neural/persist"
)

var inter float64
var bio_index []BioIndex

// 调查记录详细结果（指标值）
type BioSurveyResult struct {
	IndexId        uuid.UUID // 指标Id
	IndexType      string    // 指标类型（定量、定性单选、定性多选）
	CalculateValue float64   // 指标值
}

// 调查记录（上一结构体列表加日期）
type HitoricalBioSurvey struct {
	BioSurveyDate    time.Time         // 调查记录日期
	BioSurveyResults []BioSurveyResult // 调查记录结果
}

// 用户的调查记录列表（上一结构体列表加用户ID）
type UserHitoricalBioSurvey struct {
	UserID              uuid.UUID            // 用户Id
	HitoricalBioSurveys []HitoricalBioSurvey // 用户调查记录列表
}

// 指标信息
type BioIndex struct {
	IndexId uuid.UUID // 指标ID
	Minimum float64   // 最小值
	Maxmum  float64   // 最大值
	Weight  float64   // 权重
}

func is_need(array []float64) bool {
	// 正则化
	sort.Float64s(array)
	min_value := array[0]
	max_value := array[len(array)-1]
	for i, _ := range array {
		array[i] = (array[i] - min_value) / (max_value - min_value + 1e-3)
	}
	fmt.Println(array)
	_, stddev := stat.MeanStdDev(array, nil)
	fmt.Println(stddev)
	if stddev < 0.5 {
		return false
	}
	return true
}

// 1.计算适用指标
func CalPendingBioIndexIdList(hitoricalBioSurveys []HitoricalBioSurvey) (pendingBioIndexIdList []uuid.UUID, err error) {
	//pendingBioIndexIdList := []uuid.UUID{}
	id_value := make(map[uuid.UUID][]float64)
	//id_date := make(map[uuid.UUID][]time.Time)
	for _, i := range hitoricalBioSurveys {
		//fmt.Println(i.BioSurveyDate)
		//fmt.Println(i.BioSurveyResults)
		for _, j := range i.BioSurveyResults {
			//fmt.Println(j.IndexId)
			//fmt.Println(j.IndexType)
			//fmt.Println(j.CalculateValue)
			id_value[j.IndexId] = append(id_value[j.IndexId], j.CalculateValue)
		}
	}
	fmt.Println(id_value)
	for id, vs := range id_value {
		//fmt.Println(id) // 指标ID
		if is_need(vs) {
			pendingBioIndexIdList = append(pendingBioIndexIdList, id)
		}
	}
	return pendingBioIndexIdList, err
}

// 2.搭建模型
func CreateBioageModel(userHitoricalBioSurveys []UserHitoricalBioSurvey, bioIndex_i []BioIndex) (err error) {
// func CreateBioageModel(userHitoricalBioSurveys []UserHitoricalBioSurvey, bioIndex_i []BioIndex) (intercept float64, bioIndex_o []BioIndex, err error) {
	// 更新指标权重
	values := [][]float64{}
	//value := []float64{}
	id_value := make(map[uuid.UUID]float64) //保存指标的上一次历史值
	//num_index := len(bioIndex)
	for _, i := range userHitoricalBioSurveys {
		fmt.Println(i.UserID)
		fmt.Println(i.HitoricalBioSurveys)
		for _, j := range i.HitoricalBioSurveys {
			fmt.Println(j.BioSurveyDate)
			fmt.Println(j.BioSurveyResults)
			value := []float64{}
			for _, k := range j.BioSurveyResults {
				fmt.Println(k.IndexType)
				fmt.Println(k.IndexId)
				fmt.Println(k.CalculateValue)
				if &k.CalculateValue != nil {
					value = append(value, k.CalculateValue)
					id_value[k.IndexId] = k.CalculateValue
				} else {
					k.CalculateValue = id_value[k.IndexId]
				}
			}
			values = append(values, value)
		}
	}

	// 划分X自变量，Y因变量值
	threeDLineX := [][]float64{}
	threeDLineY := []float64{}
	col := len(values[0])
	for i, _ := range values {
		row := []float64{}
		for _, v := range values[i][:col-1] {
			row = append(row, v)
		}
		threeDLineX = append(threeDLineX, row)
		threeDLineY = append(threeDLineY, values[i][col-1])
	}

	// 线性回归模型
	// model_ := linear.NewLeastSquares(base.StochasticGA, 1e-4, 13.06, 1e1, threeDLineX, threeDLineY)
	// err = model_.Learn()
	// if err != nil {
	// 	panic("This is wrong!")
	// }
	// inter = model_.Parameters[0]
	// fmt.Println("model_.Parameters:", len(model_.Parameters))
	// fmt.Println("bio_index:", len(bioIndex_i))
	// if len(model_.Parameters)-1 != len(bioIndex_i) {
	// 	return 0, []BioIndex{}, errors.New("index num is not correct") 
	// }
	// for i, _ := range bioIndex_i {
	// 	bioIndex_i[i].Weight = model_.Parameters[i+1]
	// }
	// fmt.Println(values)
	// fmt.Println(bioIndex_i[0].Weight)
	// return intercept, bioIndex_i, err

	// 全链接神经网络
	if len(threeDLineX) != len(threeDLineY) {
		err = errors.New("input mismatch output when create model")
	}
	fmt.Println(threeDLineX)
	fmt.Println(threeDLineY)
	// epoch=1000
	speed := 0.01
	n := neural.NewNetwork(2, []int{2, 4, 1})
	n.RandomizeSynapses()
	for i := 0; i < 100; i++ {
		for j, _ := range threeDLineX {
			output := []float64{}
			input := threeDLineX[j]
			output = append(output, threeDLineY[j])
			learn.Learn(n, input, output, speed)
		}
	}
	persist.ToFile("./model.json", n)
	return err
}

//3.调用模型
func model(bioSurveyResults []BioSurveyResult) (bioage float64, err error) {
	fmt.Println(bioSurveyResults)
	fmt.Println(bio_index)
	bioage += inter
	for i, _ := range bioSurveyResults[:len(bioSurveyResults)-1] {
		bioage += bioSurveyResults[i].CalculateValue * bio_index[i].Weight
	}
	return bioage, err

	// 神经网络计算
	// v := []float64{}
	// for i, _ := range bioSurveyResults {
	// 	v = append(v, bioSurveyResults[i].CalculateValue)
	// }
	// n := persist.FromFile("./model.json")
	// fmt.Println(n)
	// fmt.Println(v)
	// fmt.Println(n.Calculate(v))
	// bioage = n.Calculate(v)[0]
	// return bioage, err
}

func main() {
	//创建指标id
	index_id1 := uuid.NewV4()
	index_id2 := uuid.NewV4()
	//创建不同指标值（2个指标，3次测量）
	bio_survey_res1 := BioSurveyResult{IndexId: index_id1, IndexType: "xing", CalculateValue: 100}
	bio_survey_res2 := BioSurveyResult{IndexId: index_id2, IndexType: "liang", CalculateValue: 4}
	bio_survey_res3 := BioSurveyResult{IndexId: index_id2, IndexType: "liang", CalculateValue: 16}
	bio_survey_res4 := BioSurveyResult{IndexId: index_id1, IndexType: "xing", CalculateValue: 80}
	bio_survey_res5 := BioSurveyResult{IndexId: index_id2, IndexType: "liang", CalculateValue: 3}
	bio_survey_res6 := BioSurveyResult{IndexId: index_id2, IndexType: "liang", CalculateValue: 17}
	bio_survey_res7 := BioSurveyResult{IndexId: index_id1, IndexType: "xing", CalculateValue: 80}
	bio_survey_res8 := BioSurveyResult{IndexId: index_id2, IndexType: "liang", CalculateValue: 2}
	bio_survey_res9 := BioSurveyResult{IndexId: index_id2, IndexType: "liang", CalculateValue: 18} // 生物年龄当作指标最后一个属性，在最后一列，用于构建模型
	// bio_survey_res3 := BioSurveyResult{IndexId: index_id2, IndexType: "liang", CalculateValue: 18} // 生物年龄当作指标最后一个属性，在最后一列，用于构建模型
	// he_bio_survey_res_1 := []BioSurveyResult{bio_survey_res1, bio_survey_res2}
	// he_bio_survey_res_2 := []BioSurveyResult{bio_survey_res3, bio_survey_res4}
	// he_bio_survey_res := []BioSurveyResult{bio_survey_res1, bio_survey_res2, bio_survey_res3}
	historical_bio_survey1 := HitoricalBioSurvey{BioSurveyDate: time.Date(2022, 4, 24, 13, 45, 0, 0, time.UTC), BioSurveyResults: []BioSurveyResult{bio_survey_res1, bio_survey_res2, bio_survey_res3}}
	historical_bio_survey2 := HitoricalBioSurvey{BioSurveyDate: time.Date(2022, 4, 24, 14, 21, 0, 0, time.UTC), BioSurveyResults: []BioSurveyResult{bio_survey_res4, bio_survey_res5, bio_survey_res6}}
	historical_bio_survey3 := HitoricalBioSurvey{BioSurveyDate: time.Date(2022, 4, 24, 15, 23, 0, 0, time.UTC), BioSurveyResults: []BioSurveyResult{bio_survey_res7, bio_survey_res8, bio_survey_res9}}
	historical_bio_survey := []HitoricalBioSurvey{historical_bio_survey1, historical_bio_survey2, historical_bio_survey3}
	index_id, err := CalPendingBioIndexIdList(historical_bio_survey)
	fmt.Println(index_id)
	fmt.Println(err)
	fmt.Println("计算适用指标结束---------------------")

	//创建用户id
	user_id1 := uuid.NewV4()
	user_id2 := uuid.NewV4()
	user_historical_bio_survey1 := []HitoricalBioSurvey{historical_bio_survey1, historical_bio_survey2, historical_bio_survey3}
	user_historical_bio_survey2 := []HitoricalBioSurvey{historical_bio_survey1, historical_bio_survey2, historical_bio_survey3}
	user_historical_bio_survey := []UserHitoricalBioSurvey{{UserID: user_id1, HitoricalBioSurveys: user_historical_bio_survey1},
		{UserID: user_id2, HitoricalBioSurveys: user_historical_bio_survey2}}
	bio_index = []BioIndex{{IndexId: index_id1, Minimum: 0, Maxmum: 100, Weight: 0}, {IndexId: index_id2, Minimum: 0, Maxmum: 100, Weight: 0}}
	inter, bio_index, err = CreateBioageModel(user_historical_bio_survey, bio_index)
	fmt.Println("构建模型结束----------------------")

	test_bio_survey_res := []BioSurveyResult{{IndexId: uuid.NewV4(), IndexType: "liang", CalculateValue: 100},
		{IndexId: uuid.NewV4(), IndexType: "xing", CalculateValue: 4}}
	bioage, _ := model(test_bio_survey_res)
	fmt.Println(bioage)
	fmt.Println("调用模型结束----------------------")
}
