package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/patrikeh/go-deep"
	"github.com/patrikeh/go-deep/training"
	uuid "github.com/satori/go.uuid"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/stat"
	"os"
	"sort"
	"time"
)

//var inter float64
var bio_index []BioIndex
var network *deep.Neural

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
	// array: 单一指标的所有历史值
	// return: True or False(历史值是否具有波动性)
	// 指标至少测量5次
	if len(array) < 5 {
		//fmt.Println("测量次数不足5次\t")
		return true
	}
	sort.Float64s(array)
	mean, stddev := stat.MeanStdDev(array, nil)
	fmt.Println(stddev)
	if stddev/mean < 0.3 {
		return false
	}
	//fmt.Println("该id变异较大\t")
	return true
}

// 1.计算适用指标(传入的是近几期的历史测量值？)
// 1.1 如果存在新增指标还没有记录时，需要判断
// 1.2 测量次数至少为5次
// 1.3 指标未测量达2个月需重新测量
// 1.4 如果全指标的个数减少，输出的适应指标相应减少
// 1.5 使用变异系数作为测量指标，不需要标准化
func CalPendingBioIndexIdList(hitoricalBioSurveys []HitoricalBioSurvey, bioindex []BioIndex) (pendingBioIndexIdList []uuid.UUID, err error) {
	//func CalPendingBioIndexIdList(currentDate time.Time, hitoricalBioSurveys []HitoricalBioSurvey, bioindex []BioIndex) (pendingBioIndexIdList map[uuid.UUID]int, err error) {
	id_value := make(map[uuid.UUID][]float64) //key:指标ID;value:历次指标值
	id_date := make(map[uuid.UUID]time.Time)  //记录每个指标的最新测量时间
	bio_num := len(hitoricalBioSurveys)
	for i := 0; i < bio_num; i++ {
		//fmt.Println(i.BioSurveyDate)
		//fmt.Println(i.BioSurveyResults)
		for _, j := range hitoricalBioSurveys[i].BioSurveyResults {
			//fmt.Println(j.IndexId)
			//fmt.Println(j.IndexType)
			//fmt.Println(j.CalculateValue)
			id_value[j.IndexId] = append(id_value[j.IndexId], j.CalculateValue)
			id_date[j.IndexId] = hitoricalBioSurveys[i].BioSurveyDate
		}
	}

	// ADD: 1.3 指标未测量达2个月需重新测量
	for i, t := range id_date {
		if time.Now().Sub(t).Hours()/24 > 60 {
			fmt.Println("测量时间间隔超过2个月，添加该id", i)
			pendingBioIndexIdList = append(pendingBioIndexIdList, i)
		} else {
			if is_need(id_value[i]) {
				fmt.Println("遂添加该id", i)
				pendingBioIndexIdList = append(pendingBioIndexIdList, i)
			} else {
				fmt.Println("This id dont need mearure!", i)
			}
		}
	}

	// ADD: 1.1 如果存在新增指标还没有记录时，需要判断
	fmt.Println(id_value)
	for _, bi := range bioindex {
		_, ok := id_value[bi.IndexId]
		if ok {
			//fmt.Println("This index exists!, indexHistoryValue:", vs)
		} else {
			//fmt.Println("This index not exists！")
			pendingBioIndexIdList = append(pendingBioIndexIdList, bi.IndexId)
		}
	}

	// 1.4 如果全指标的个数减少，输出的适应指标相应减少
	for i := 0; i < len(pendingBioIndexIdList); i++ {
		flags := false
		for j, _ := range bioindex {
			if pendingBioIndexIdList[i] == bioindex[j].IndexId {
				flags = true
				break
			}
		}
		if !flags {
			if i != len(pendingBioIndexIdList)-1 {
				pendingBioIndexIdList = append(pendingBioIndexIdList[:i], pendingBioIndexIdList[i+1:]...)
			} else {
				pendingBioIndexIdList = pendingBioIndexIdList[:i]
			}
			i -= 1
		}
	}

	//for id, vs := range id_value {
	//fmt.Println(id) // 指标ID
	//if is_need(vs) {
	//	if is_chong(pendingBioIndexIdList, id) {
	//		fmt.Println("This id is alreagy exists!")
	//	} else {
	//		pendingBioIndexIdList = append(pendingBioIndexIdList, id)
	//	}
	//}
	//}
	return pendingBioIndexIdList, err
}

func is_contain(a []uuid.UUID, b []uuid.UUID) bool {
	// 判断a数组是否包含b数组
	//if len(a) != len(b) {
	//	return false
	//}
	//for i, _ := range a {
	//	if a[i] != b[i] {
	//		return false
	//	}
	//}
	//return true
	sum := 0
	for i, _ := range b {
		for j, _ := range a {
			if b[i] == a[j] {
				sum += 1
				break
			}
		}
	}
	if sum == len(b) {
		return true
	}
	return false
}

func is_exist(array uuid.UUID, arrays []uuid.UUID) bool {
	for _, uid := range arrays {
		if uid == array {
			return true
		}
	}
	return false
}

// 2.搭建模型
// 2.1 添加全指标，以补全未测量指标部分的值
// 2.2 模型保存为文件
// 2.3 最新的全指标还未测量，返回error，告知用户需先进行测量
// 2.4 构建模型前进行归一化
func CreateBioageModel(userHitoricalBioSurveys []UserHitoricalBioSurvey, bioIndex []BioIndex) (n_out *deep.Neural, err error) {
	//用户记录安装日期顺序升序
	//func CreateBioageModel(userHitoricalBioSurveys []UserHitoricalBioSurvey, bioIndex_i []BioIndex) (err error) {
	//func CreateBioageModel(userHitoricalBioSurveys []UserHitoricalBioSurvey, bioIndex_i []BioIndex) (intercept float64, bioIndex_o []BioIndex, err error) {
	if len(userHitoricalBioSurveys) == 0 {
		return n_out, errors.New("所传用户记录为空，无法构建模型")
	}
	if len(bioIndex) == 0 {
		return n_out, errors.New("所传全指标信息为空，无法构建模型")
	}
	//	ADD: 2.1 添加全指标，以补全未测量指标部分的值
	values := [][]float64{}
	to_indids := []uuid.UUID{}
	for _, b := range bioIndex {
		to_indids = append(to_indids, b.IndexId)
	}
	for _, i := range userHitoricalBioSurveys {
		var index_num int = -1 // 每个用户开始采用记录构建模型的起始位置
		cur_indids := []uuid.UUID{}
		id_value := make(map[uuid.UUID]float64) // 保存每个用户指标的起始位置的历史值
		for pos, j := range i.HitoricalBioSurveys {
			//value := []float64{}
			for _, k := range j.BioSurveyResults {
				if !is_exist(k.IndexId, cur_indids) {
					cur_indids = append(cur_indids, k.IndexId)
				}
				id_value[k.IndexId] = k.CalculateValue
			}
			if is_contain(cur_indids, to_indids) {
				index_num = pos
				break
			} else {
				continue
			}
		}
		if index_num == -1 {
			continue
		}
		for _, j := range i.HitoricalBioSurveys[index_num:] {
			value := []float64{}
			for _, k := range j.BioSurveyResults {
				id_value[k.IndexId] = k.CalculateValue //指标对应值更新为最新日期的
			}
			for _, m := range bioIndex {
				value = append(value, id_value[m.IndexId])
			}
			values = append(values, value)
		}
	}

	//for _, i := range userHitoricalBioSurveys {
	//	_, ok := index_num[i.UserID]
	//	if ok {
	//		if len(i.HitoricalBioSurveys)-index_num[i.UserID] < 10 {
	//			return nil, errors.New("最新应用的测量指标次数不超过10次")
	//		} else {
	//			for _, j := range i.HitoricalBioSurveys[index_num[i.UserID]:] {
	//				value := []float64{}
	//				cur_ids := []uuid.UUID{}
	//				for _, k := range j.BioSurveyResults {
	//					cur_ids = append(cur_ids, k.IndexId)
	//				}
	//				for _, m := range ids {
	//					}
	//					for _, m := range ids {
	//						if k.IndexId == m {
	//							value = append(value, k.CalculateValue)
	//						}
	//						//fmt.Println(k.IndexType)
	//						//fmt.Println(k.IndexId)
	//						//fmt.Println(k.CalculateValue)
	//						//若指标未测量，则根据指标以往最近一次历史值补全
	//						//若指标是空值，则根据指标以往最近一次历史值补全
	//						if &k.CalculateValue != nil {
	//							value = append(value, k.CalculateValue)
	//							id_value[k.IndexId] = k.CalculateValue
	//						} else {
	//							//k.CalculateValue = id_value[k.IndexId]
	//							value = append(value, id_value[k.IndexId])
	//						}
	//					}
	//				}
	//				values = append(values, value)
	//			}
	//		}
	//}

	//num_index := len(bioIndex)
	//for _, i := range userHitoricalBioSurveys {
	//	for _, j := range i.HitoricalBioSurveys {
	//		value := []float64{}
	//		for _, k := range j.BioSurveyResults {
	//			//fmt.Println(k.IndexType)
	//			//fmt.Println(k.IndexId)
	//			//fmt.Println(k.CalculateValue)
	//			//若指标未测量或是空值，则根据指标以往最近一次历史值补全
	//			if &k.CalculateValue != nil {
	//				value = append(value, k.CalculateValue)
	//				id_value[k.IndexId] = k.CalculateValue
	//			} else {
	//				//k.CalculateValue = id_value[k.IndexId]
	//				value = append(value, id_value[k.IndexId])
	//			}
	//		}
	//		values = append(values, value)
	//	}
	//}

	//ADD：2.4 构建模型前进行归一化
	row := len(values)
	col := len(values[0])
	values_T := make([][]float64, col, col+1)
	for i, _ := range values {
		for j := 0; j < col; j++ {
			values_T[j] = append(values_T[j], values[i][j])
		}
	}
	for i := 0; i < len(values_T)-1; i++ {
		min_value := floats.Min(values_T[i])
		max_value := floats.Max(values_T[i])
		bio_index[i].Minimum = min_value
		bio_index[i].Maxmum = max_value
		for j, _ := range values_T[i] {
			values_T[i][j] = (values_T[i][j] - min_value) / (max_value - min_value)
		}
	}
	values_nor := make([][]float64, row, row+1)
	for i, _ := range values_T {
		for j, _ := range values_T[i] {
			values_nor[j] = append(values_nor[j], values_T[i][j])
		}
	}
	fmt.Println(values_nor)

	//ADD: 2.3 最新的全指标还未测量，返回error，告知用户需先进行测量
	if len(values_nor) == 0 {
		return n_out, errors.New("ERROR: 请先根据最新指标列表进行记录的测量")
	}
	if len(values_nor[0]) < 2 {
		return n_out, errors.New("ERROR: 输入数据维度为1，无法划分自变量和因变量")
	}

	// 划分X自变量，Y因变量值,默认Y排在指标最后
	threeDLineX := [][]float64{}
	threeDLineY := []float64{}
	col = len(values_nor[0])
	for i, _ := range values_nor {
		feature := []float64{}
		for _, v := range values_nor[i][:col-1] {
			feature = append(feature, v)
		}
		threeDLineX = append(threeDLineX, feature)
		threeDLineY = append(threeDLineY, values_nor[i][col-1])
	}

	// 线性回归模型
	//model_ := linear.NewLeastSquares(base.StochasticGA, 1e-4, 13.06, 1e1, threeDLineX, threeDLineY)
	//err = model_.Learn()
	//if err != nil {
	//	panic("This is wrong!")
	//}
	//inter = model_.Parameters[0]
	//fmt.Println("model_.Parameters:", len(model_.Parameters))
	//fmt.Println("bio_index:", len(bioIndex_i))
	//if len(model_.Parameters)-1 != len(bioIndex_i) {
	//	return 0, []BioIndex{}, errors.New("index num is not correct")
	//}
	//for i, _ := range bioIndex_i {
	//	bioIndex_i[i].Weight = model_.Parameters[i+1]
	//}
	//fmt.Println(values)
	//fmt.Println(bioIndex_i[0].Weight)
	//return intercept, bioIndex_i, err

	// 全链接神经网络go-neural
	//if len(threeDLineX) != len(threeDLineY) {
	//	err = errors.New("input mismatch output when create model")
	//}
	//epoch := 160
	//speed := 0.1
	//n := neural.NewNetwork(2, []int{10, 1})
	//n.RandomizeSynapses()
	//for i := 0; i < epoch; i++ {
	//	for j, _ := range threeDLineX {
	//		output := []float64{}
	//		input := threeDLineX[j]
	//		output = append(output, threeDLineY[j])
	//		learn.Learn(n, input, output, speed)
	//	}
	//}
	//persist.ToFile("./model.json", n)
	//return err

	// 全链接神经网络go-deep
	if len(threeDLineX) != len(threeDLineY) {
		return n_out, errors.New("input mismatch output when create model")
	}
	data := training.Examples{}
	for i, _ := range threeDLineX {
		data = append(data, training.Example{Input: threeDLineX[i], Response: []float64{threeDLineY[i]}})
	}
	n := deep.NewNeural(&deep.Config{
		/* Input dimensionality */
		Inputs: len(values[0]) - 1,
		/* Two hidden layers consisting of two neurons each, and a single output */
		Layout: []int{1, 1},
		/* Activation functions: Sigmoid, Tanh, ReLU, Linear */
		Activation: deep.ActivationNone,
		/* Determines output layer activation & loss function:
		   ModeRegression: linear outputs with MSE loss
		   ModeMultiClass: softmax output with Cross Entropy loss
		   ModeMultiLabel: sigmoid output with Cross Entropy loss
		   ModeBinary: sigmoid output with binary CE loss */
		Mode: deep.ModeRegression,
		/* Weight initializers: {deep.NewNormal(μ, σ), deep.NewUniform(μ, σ)} */
		Weight: deep.NewNormal(1.0, 0.0),
		/* Apply bias */
		Bias: true,
	})
	// params: learning rate, momentum, alpha decay, nesterov
	optimizer := training.NewSGD(0.05, 0.1, 1e-2, true)
	// params: optimizer, verbosity (print stats at every 50th iteration)
	trainer := training.NewTrainer(optimizer, 50)
	// start training
	training, heldout := data.Split(1)
	trainer.Train(n, training, heldout, 100) // training, validation, iterations
	//fmt.Printf("n:%T\n", n)
	//fmt.Println("n", n)
	//fmt.Println(n.Predict([]float64{100, 4}))
	//persist.ToFile("./model.json", n)

	// ADD: 2.2 模型保存为文件
	dump, _ := json.Marshal(n.Dump())
	os.WriteFile("./model.json", dump, 0777)

	return n, err
}

//3.调用模型
//3.1更改传入参数以补全最新记录的缺失值，参照构建模型时的预处理
//3.2读取本地模型文件
//3.3最新的全指标还未测量，返回error，告知用户需先进行测量
//3.4调用模型前进行归一化
func model(hitBioSurveyResults []HitoricalBioSurvey, bioIndex []BioIndex) (bioage float64, err error) {
	//bioIndex: 相比于构建模型应当减去Y（生物年龄）列

	// 神经网络计算go-neural
	//v := []float64{}
	//for i, _ := range bioSurveyResults {
	//	v = append(v, bioSurveyResults[i].CalculateValue)
	//}
	//n := persist.FromFile("./model.json")
	//fmt.Println(n)
	//fmt.Println(v)
	//fmt.Println(n.Calculate(v))
	//bioage = n.Calculate(v)[0]

	if len(hitBioSurveyResults) == 0 {
		return 0, errors.New("所传测量的历史记录为空，无法调用模型")
	}
	if len(bioIndex) == 0 {
		return 0, errors.New("所传全指标信息为空，无法调用模型")
	}

	//神经网络计算go-deep
	//ADD: 3.1更改传入参数以补全最新记录的缺失值，参照构建模型时的预处理
	num_bio := len(hitBioSurveyResults)
	id_value := make(map[uuid.UUID]float64) //保存各指标的历史值
	to_indids := []uuid.UUID{}
	var index_num int = -1 //记录所有指标齐全的开始日期位置索引
	for _, b := range bioIndex {
		to_indids = append(to_indids, b.IndexId)
	}
	cur_indids := []uuid.UUID{}
	for pos, j := range hitBioSurveyResults {
		for _, k := range j.BioSurveyResults {
			if !is_exist(k.IndexId, cur_indids) {
				cur_indids = append(cur_indids, k.IndexId)
			}
			id_value[k.IndexId] = k.CalculateValue
		}
		if is_contain(cur_indids, to_indids) {
			index_num = pos
			//for _, k := range j.BioSurveyResults {
			//	id_value[k.IndexId] = k.CalculateValue
			//value = append(value, k.CalculateValue)
			//}
			break
		} else {
			continue
		}
		//values = append(values, value)
	}
	fmt.Println("index_num:", index_num)
	if index_num == -1 {
		return 0, errors.New("当前全指标记录还未测量，请测量并构建模型后再调用")
	}
	for _, i := range hitBioSurveyResults[index_num:] {
		for _, j := range i.BioSurveyResults {
			id_value[j.IndexId] = j.CalculateValue
		}
	}
	//ADD: 3.3 最新的全指标还未测量，返回error，告知用户需先进行测量
	//if len(id_value) == 0 {
	//	return bioage, errors.New("ERROR: 请先根据最新指标列表进行记录的测量")
	//}
	//ids1 := make(map[uuid.UUID]float64)
	v := []float64{}
	latest_bio := hitBioSurveyResults[num_bio-1].BioSurveyResults
	for _, k := range latest_bio {
		id_value[k.IndexId] = k.CalculateValue
	}
	for _, m := range bioIndex {
		_, ok := id_value[m.IndexId]
		if ok {
			v = append(v, id_value[m.IndexId])
		} else {
			//v = append(v, id_value[m.IndexId])
			return 0, errors.New("ERROR: 请先根据最新指标列表进行记录的测量")
		}
	}
	fmt.Println("v:", v)

	//ADD：3.4调用模型前进行归一化
	//todo work

	//ADD：3.2读取本地模型文件
	undump, _ := os.ReadFile("./model.json")
	var new_dump deep.Dump
	err = json.Unmarshal(undump, &new_dump)
	new_n := deep.FromDump(&new_dump)
	//fmt.Println(new_n.Config.Inputs)
	if len(v) != new_n.Config.Inputs {
		return 0, errors.New("最新测量指标与模型输入维度不符，请检查后再调用模型")
	}
	//new_n.Forward(v[:len(v)-1])
	bioage = new_n.Predict(v)[0]
	return bioage, err
}

func main() {
	//创建指标id
	index_id1 := uuid.NewV4()
	fmt.Println(index_id1)
	index_id2 := uuid.NewV4()
	fmt.Println(index_id2)
	index_id3 := uuid.NewV4()
	fmt.Println(index_id3)
	//创建不同指标值（2个x指标，1个y指标，3次测量）
	bio_survey_res1 := BioSurveyResult{IndexId: index_id1, IndexType: "xing", CalculateValue: 1}
	bio_survey_res2 := BioSurveyResult{IndexId: index_id2, IndexType: "liang", CalculateValue: 2}
	bio_survey_res3 := BioSurveyResult{IndexId: index_id3, IndexType: "liang", CalculateValue: 26} // 生物年龄当作指标最后一个属性，在最后一列，用于构建模型
	bio_survey_res4 := BioSurveyResult{IndexId: index_id1, IndexType: "xing", CalculateValue: 2}
	bio_survey_res5 := BioSurveyResult{IndexId: index_id2, IndexType: "liang", CalculateValue: 3}
	bio_survey_res6 := BioSurveyResult{IndexId: index_id3, IndexType: "liang", CalculateValue: 27}
	bio_survey_res7 := BioSurveyResult{IndexId: index_id1, IndexType: "xing", CalculateValue: 1}
	bio_survey_res8 := BioSurveyResult{IndexId: index_id2, IndexType: "liang", CalculateValue: 2}
	bio_survey_res9 := BioSurveyResult{IndexId: index_id3, IndexType: "liang", CalculateValue: 28}
	//bio_survey_res10 := BioSurveyResult{IndexId: uuid.NewV4(), IndexType: "liang", CalculateValue: 18}
	historical_bio_survey1 := HitoricalBioSurvey{BioSurveyDate: time.Date(2022, 2, 24, 13, 45, 0, 0, time.UTC), BioSurveyResults: []BioSurveyResult{bio_survey_res1, bio_survey_res2, bio_survey_res3}}
	historical_bio_survey2 := HitoricalBioSurvey{BioSurveyDate: time.Date(2022, 3, 24, 14, 21, 0, 0, time.UTC), BioSurveyResults: []BioSurveyResult{bio_survey_res4, bio_survey_res5, bio_survey_res6}}
	historical_bio_survey3 := HitoricalBioSurvey{BioSurveyDate: time.Date(2022, 4, 24, 15, 23, 0, 0, time.UTC), BioSurveyResults: []BioSurveyResult{bio_survey_res7, bio_survey_res8, bio_survey_res9}}
	historical_bio_survey := []HitoricalBioSurvey{historical_bio_survey1, historical_bio_survey2, historical_bio_survey3}
	bio_index = []BioIndex{{IndexId: index_id1, Minimum: 0, Maxmum: 100, Weight: 0}, {IndexId: index_id2, Minimum: 0, Maxmum: 100, Weight: 0}, {IndexId: index_id3, Minimum: 0, Maxmum: 100, Weight: 0}}
	index_id, err := CalPendingBioIndexIdList(historical_bio_survey, bio_index)
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
	//inter, bio_index, err = CreateBioageModel(user_historical_bio_survey, bio_index)
	//err = CreateBioageModel(user_historical_bio_survey, bio_index)
	network, err = CreateBioageModel(user_historical_bio_survey, bio_index)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("构建模型结束----------------------")

	//bioage, _ := model(test_bio_survey_res)
	//new_bio_index := []BioIndex{{IndexId: uuid.FromStringOrNil("2ff020c7-5cd9-4340-a622-ea7c6d78b6a8"), Minimum: 0, Maxmum: 100, Weight: 0}, {IndexId: uuid.FromStringOrNil("9aaba953-f366-441c-9b95-6dca0df9f780"), Minimum: 0, Maxmum: 100, Weight: 0}, {IndexId: uuid.FromStringOrNil("0292093c-3544-4f04-855a-76bdd467d195"), Minimum: 0, Maxmum: 100, Weight: 0}, {IndexId: uuid.FromStringOrNil("8dcf9565-5f9b-49f8-9c3f-c7b2a5f6206b"), Minimum: 0, Maxmum: 100, Weight: 0}, {IndexId: uuid.FromStringOrNil("c348ed0a-9892-44b8-9cb7-836af9deb9f9"), Minimum: 0, Maxmum: 100, Weight: 0}}
	new_bio_index := []BioIndex{{IndexId: index_id1, Minimum: 0, Maxmum: 100, Weight: 0}, {IndexId: index_id2, Minimum: 0, Maxmum: 100, Weight: 0}}
	//new_bio_survey_res1 := BioSurveyResult{IndexId: uuid.FromStringOrNil("2ff020c7-5cd9-4340-a622-ea7c6d78b6a8"), IndexType: "xing", CalculateValue: 2}
	//new_bio_survey_res2 := BioSurveyResult{IndexId: uuid.FromStringOrNil("9aaba953-f366-441c-9b95-6dca0df9f780"), IndexType: "xing", CalculateValue: 30}
	//new_bio_survey_res3 := BioSurveyResult{IndexId: uuid.FromStringOrNil("0292093c-3544-4f04-855a-76bdd467d195"), IndexType: "xing", CalculateValue: 2}
	//new_bio_survey_res4 := BioSurveyResult{IndexId: uuid.FromStringOrNil("8dcf9565-5f9b-49f8-9c3f-c7b2a5f6206b"), IndexType: "xing", CalculateValue: 2}
	//new_bio_survey_res5 := BioSurveyResult{IndexId: uuid.FromStringOrNil("c348ed0a-9892-44b8-9cb7-836af9deb9f9"), IndexType: "xing", CalculateValue: 17}
	//new_bio_survey_res6 := BioSurveyResult{IndexId: uuid.FromStringOrNil("b6eeda60-8150-46fd-8d48-a6adae7893fd"), IndexType: "xing", CalculateValue: 3}
	new_historical_bio_survey1 := HitoricalBioSurvey{BioSurveyDate: time.Date(2022, 2, 24, 13, 45, 0, 0, time.UTC), BioSurveyResults: []BioSurveyResult{bio_survey_res1, bio_survey_res2}}
	new_historical_bio_survey2 := HitoricalBioSurvey{BioSurveyDate: time.Date(2022, 3, 24, 14, 21, 0, 0, time.UTC), BioSurveyResults: []BioSurveyResult{bio_survey_res4, bio_survey_res5}}
	new_historical_bio_survey3 := HitoricalBioSurvey{BioSurveyDate: time.Date(2022, 4, 24, 15, 23, 0, 0, time.UTC), BioSurveyResults: []BioSurveyResult{bio_survey_res7, bio_survey_res8}}
	new_user_historical_bio_survey := []HitoricalBioSurvey{new_historical_bio_survey1, new_historical_bio_survey2, new_historical_bio_survey3}
	bioage, err := model(new_user_historical_bio_survey, new_bio_index)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(bioage)
	fmt.Println("调用模型结束----------------------")
}
