package main

import (
	"errors"
	"fmt"
	"github.com/cdipaolo/goml/base"
	"github.com/cdipaolo/goml/linear"
	"gonum.org/v1/gonum/stat"
	"sort"

	//"github.com/grd/statistics"
	uuid "github.com/satori/go.uuid"
	"time"
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
func CreateBioageModel(userHitoricalBioSurveys []UserHitoricalBioSurvey, bioIndex_i []BioIndex) (intercept float64, bioIndex_o []BioIndex, err error) {
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
	model_ := linear.NewLeastSquares(base.StochasticGA, 1e-4, 13.06, 1e3, threeDLineX, threeDLineY)
	err = model_.Learn()
	if err != nil {
		panic("This is wrong!")
	}
	intercept = model_.Parameters[0]
	if len(model_.Parameters)-1 != len(bioIndex_i) {
		return 0, []BioIndex{}, errors.New("index num is not correct")
	}
	for i, _ := range bioIndex_i {
		bioIndex_i[i].Weight = model_.Parameters[i+1]
	}
	fmt.Println(values)
	fmt.Println(bioIndex_i[0].Weight)
	return intercept, bioIndex_i, err
}

//3.调用模型
func model(bioSurveyResults []BioSurveyResult) (bioage float64, err error) {
	bioage += inter
	for i, _ := range bioSurveyResults[:len(bioSurveyResults)-1] {
		bioage += bioSurveyResults[i].CalculateValue * bio_index[i].Weight
	}
	return bioage, err
}

func main() {
	index_u1 := uuid.NewV4()
	index_u2 := uuid.NewV4()
	user_u1 := uuid.NewV4()
	user_u2 := uuid.NewV4()
	bio_survey_res1 := BioSurveyResult{IndexId: index_u1, IndexType: "xing", CalculateValue: 100}
	bio_survey_res2 := BioSurveyResult{IndexId: index_u2, IndexType: "liang", CalculateValue: 4}
	//bio_survey_res3 := BioSurveyResult{IndexId: index_u1, IndexType: "xing", CalculateValue: 90}
	//bio_survey_res4 := BioSurveyResult{IndexId: index_u2, IndexType: "liang", CalculateValue: 3}
	bio_survey_res3 := BioSurveyResult{IndexId: index_u2, IndexType: "liang", CalculateValue: 18} // 生物年龄当作指标最后一个属性，在最后一列
	//he_bio_survey_res_1 := []BioSurveyResult{bio_survey_res1, bio_survey_res2}
	//he_bio_survey_res_2 := []BioSurveyResult{bio_survey_res3, bio_survey_res4}
	he_bio_survey_res := []BioSurveyResult{bio_survey_res1, bio_survey_res2, bio_survey_res3}
	historical_bio_survey1 := HitoricalBioSurvey{BioSurveyDate: time.Date(2022, 4, 24, 13, 45, 0, 0, time.UTC), BioSurveyResults: he_bio_survey_res}
	historical_bio_survey2 := HitoricalBioSurvey{BioSurveyDate: time.Date(2022, 4, 24, 14, 21, 0, 0, time.UTC), BioSurveyResults: he_bio_survey_res}
	historical_bio_survey3 := HitoricalBioSurvey{BioSurveyDate: time.Date(2022, 4, 24, 15, 23, 0, 0, time.UTC), BioSurveyResults: he_bio_survey_res}
	//historical_bio_survey := []HitoricalBioSurvey{historical_bio_survey1, historical_bio_survey2, historical_bio_survey3}
	//index_id, err := CalPendingBioIndexIdList(historical_bio_survey)
	//fmt.Println(index_id)
	//fmt.Println(err)

	user_historical_bio_survey1 := []HitoricalBioSurvey{historical_bio_survey1, historical_bio_survey2, historical_bio_survey3}
	user_historical_bio_survey2 := []HitoricalBioSurvey{historical_bio_survey1, historical_bio_survey2, historical_bio_survey3}
	user_historical_bio_survey := []UserHitoricalBioSurvey{{UserID: user_u1, HitoricalBioSurveys: user_historical_bio_survey1},
		{UserID: user_u2, HitoricalBioSurveys: user_historical_bio_survey2}}
	bio_index = []BioIndex{{IndexId: index_u1, Minimum: 0, Maxmum: 100, Weight: 0}, {IndexId: index_u2, Minimum: 0, Maxmum: 100, Weight: 0}}
	inter, bio_index, _ = CreateBioageModel(user_historical_bio_survey, bio_index)

	bioage, _ := model(he_bio_survey_res)
	fmt.Println(bioage)
}
